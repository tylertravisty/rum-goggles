package chatbot

import (
	"bufio"
	"cmp"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tylertravisty/rum-goggles/v1/internal/models"
)

const (
	PrefixAccount = "/user/"
	PrefixChannel = "/c/"
)

func SortRules(rules []Rule) {
	slices.SortFunc(rules, func(a, b Rule) int {
		return cmp.Compare(strings.ToLower(a.Display), strings.ToLower(b.Display))
	})
}

type Rule struct {
	ID         *int64          `json:"id"`
	ChatbotID  *int64          `json:"chatbot_id"`
	Display    string          `json:"display"`
	Parameters *RuleParameters `json:"parameters"`
	Running    bool            `json:"running"`
}

type Page struct {
	Name   string
	Prefix string
}

func (r *Rule) Page() *Page {
	if r.Parameters != nil {
		return r.Parameters.Page()
	}

	return nil
}

type RuleParameters struct {
	Message *RuleMessage `json:"message"`
	SendAs  *RuleSender  `json:"send_as"`
	Trigger *RuleTrigger `json:"trigger"`
}

func (rp *RuleParameters) Page() *Page {
	if rp.Trigger != nil {
		return rp.Trigger.Page()
	}

	return nil
}

type RuleMessage struct {
	FromFile *RuleMessageFile `json:"from_file"`
	FromText string           `json:"from_text"`
}

func (rm *RuleMessage) String() (string, error) {
	if rm.FromFile == nil {
		return rm.FromText, nil
	}

	s, err := rm.FromFile.string()
	if err != nil {
		return "", fmt.Errorf("error reading from file: %v", err)
	}

	return s, nil
}

func (rmf *RuleMessageFile) string() (string, error) {
	if rmf.Filepath == "" {
		return "", fmt.Errorf("filepath is empty")
	}

	if len(rmf.lines) == 0 {
		file, err := os.Open(rmf.Filepath)
		if err != nil {
			return "", fmt.Errorf("error opening file: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			rmf.lines = append(rmf.lines, line)
		}

		if len(rmf.lines) == 0 {
			return "", fmt.Errorf("no lines read")
		}
	}

	if rmf.RandomRead {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(rmf.lines))))
		if err != nil {
			return "", fmt.Errorf("error generating random line number: %v", err)
		}

		return rmf.lines[n.Int64()], nil
	}

	line := rmf.lines[rmf.lineNum]
	rmf.lineNum = rmf.lineNum + 1
	if rmf.lineNum >= len(rmf.lines) {
		rmf.lineNum = 0
	}

	return line, nil
}

type RuleMessageFile struct {
	Filepath   string `json:"filepath"`
	RandomRead bool   `json:"random_read"`
	lines      []string
	lineNum    int
}

type RuleSender struct {
	ChannelID *string `json:"channel_id"`
	Display   string  `json:"display"`
	Username  string  `json:"username"`
}

func (rs *RuleSender) ChannelIDInt() (*int, error) {
	if rs.ChannelID == nil {
		return nil, nil
	}

	i64, err := strconv.ParseInt(*rs.ChannelID, 10, 64)
	if err != nil {
		return nil, pkgErr("error parsing channel ID", err)
	}
	i := int(i64)

	return &i, nil
}

type RuleTrigger struct {
	OnCommand *RuleTriggerCommand `json:"on_command"`
	OnEvent   *RuleTriggerEvent   `json:"on_event"`
	OnTimer   *time.Duration      `json:"on_timer"`
}

func (rt *RuleTrigger) Page() *Page {
	if rt.OnEvent != nil {
		return rt.OnEvent.Page()
	}

	return nil
}

type RuleTriggerCommand struct {
	Command  string                         `json:"command"`
	Restrict *RuleTriggerCommandRestriction `json:"restrict"`
	Timeout  time.Duration                  `json:"timeout"`
}

type RuleTriggerCommandRestriction struct {
	Bypass       *RuleTriggerCommandRestrictionBypass `json:"bypass"`
	ToAdmin      bool                                 `json:"to_admin"`
	ToFollower   bool                                 `json:"to_follower"`
	ToMod        bool                                 `json:"to_mod"`
	ToStreamer   bool                                 `json:"to_streamer"`
	ToSubscriber bool                                 `json:"to_subscriber"`
	ToRant       int                                  `json:"to_rant"`
}

type RuleTriggerCommandRestrictionBypass struct {
	IfAdmin    bool `json:"if_admin"`
	IfMod      bool `json:"if_mod"`
	IfStreamer bool `json:"if_streamer"`
}

type RuleTriggerEvent struct {
	FromAccount    *RuleTriggerEventAccount    `json:"from_account"`
	FromChannel    *RuleTriggerEventChannel    `json:"from_channel"`
	FromLiveStream *RuleTriggerEventLiveStream `json:"from_live_stream"`
}

func (rte *RuleTriggerEvent) Page() *Page {
	switch {
	case rte.FromAccount != nil:
		return &Page{rte.FromAccount.Name, PrefixAccount}
	case rte.FromChannel != nil:
		return &Page{rte.FromChannel.Name, PrefixChannel}
	default:
		return nil
	}
}

type RuleTriggerEventAccount struct {
	Name     string                         `json:"name"`
	OnFollow *RuleTriggerEventAccountFollow `json:"on_follow"`
}

type RuleTriggerEventAccountFollow struct{}

type RuleTriggerEventChannel struct {
	Name     string                         `json:"name"`
	OnFollow *RuleTriggerEventChannelFollow `json:"on_follow"`
}

type RuleTriggerEventChannelFollow struct{}

type RuleTriggerEventLiveStream struct {
	OnRaid *RuleTriggerEventLiveStreamRaid `json:"on_raid"`
	OnRant *RuleTriggerEventLiveStreamRant `json:"on_rant"`
	OnSub  *RuleTriggerEventLiveStreamSub  `json:"on_sub"`
}

type RuleTriggerEventLiveStreamRaid struct{}
type RuleTriggerEventLiveStreamRant struct {
	MinAmount int `json:"min_amount"`
	MaxAmount int `json:"max_amount"`
}
type RuleTriggerEventLiveStreamSub struct{}

func (rule *Rule) ToModelsChatbotRule() (*models.ChatbotRule, error) {
	modelsRule := &models.ChatbotRule{
		ID:        rule.ID,
		ChatbotID: rule.ChatbotID,
	}

	if rule.Parameters != nil {
		paramsB, err := json.Marshal(rule.Parameters)
		if err != nil {
			return nil, fmt.Errorf("error marshaling parameters into json: %v", err)
		}

		paramsS := string(paramsB)
		modelsRule.Parameters = &paramsS
	}

	return modelsRule, nil
}
