package chatbot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/v1/internal/events"
	"github.com/tylertravisty/rum-goggles/v1/internal/models"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type user struct {
	livestreams   map[string]*rumblelivestreamlib.Client
	livestreamsMu sync.Mutex
}

func (u *user) byLivestream(url string) *rumblelivestreamlib.Client {
	client, _ := u.livestreams[url]

	return client
}

type clients map[string]*user

func (c clients) byUsername(username string) *user {
	user, _ := c[username]
	return user
}

func (c clients) byUsernameLivestream(username string, url string) *rumblelivestreamlib.Client {
	user := c.byUsername(username)
	if user == nil {
		return nil
	}

	return user.byLivestream(url)
}

type receiver struct {
	onCommand   map[string]map[int64]chan events.Chat
	onCommandMu sync.Mutex
	//onFollow []chan ???
	//onRant      []chan events.Chat
	//onSubscribe []chan events.Chat
}

type Bot struct {
	runners   map[int64]*Runner
	runnersMu sync.Mutex
}

type Chatbot struct {
	accountS    models.AccountService
	bots        map[int64]*Bot
	botsMu      sync.Mutex
	chatbotS    models.ChatbotService
	clients     clients
	clientsMu   sync.Mutex
	logError    *log.Logger
	receivers   map[string]*receiver
	receiversMu sync.Mutex
	//runners     map[int64]*Runner
	// runnersMu sync.Mutex
	wails context.Context
}

func New(accountS models.AccountService, chatbotS models.ChatbotService, logError *log.Logger, wails context.Context) *Chatbot {
	return &Chatbot{
		accountS:  accountS,
		bots:      map[int64]*Bot{},
		chatbotS:  chatbotS,
		clients:   map[string]*user{},
		logError:  logError,
		receivers: map[string]*receiver{},
		// runners:   map[int64]*Runner{},
		wails: wails,
	}
}

// TODO: resetClient/updateClient
func (cb *Chatbot) addClient(username string, livestreamUrl string) (*rumblelivestreamlib.Client, error) {
	cb.clientsMu.Lock()
	defer cb.clientsMu.Unlock()

	u := cb.clients.byUsername(username)
	if u == nil {
		u = &user{
			livestreams: map[string]*rumblelivestreamlib.Client{},
		}
		cb.clients[username] = u
	}

	client := u.byLivestream(livestreamUrl)
	if client != nil {
		return client, nil
	}

	account, err := cb.accountS.ByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("error querying account by username: %v", err)
	}

	var cookies []*http.Cookie
	err = json.Unmarshal([]byte(*account.Cookies), &cookies)
	if err != nil {
		return nil, fmt.Errorf("error un-marshaling cookie string: %v", err)
	}
	client, err = rumblelivestreamlib.NewClient(rumblelivestreamlib.NewClientOptions{Cookies: cookies, LiveStreamUrl: livestreamUrl})
	if err != nil {
		return nil, fmt.Errorf("error creating new client: %v", err)
	}

	_, err = client.ChatInfo(true)
	if err != nil {
		return nil, fmt.Errorf("error getting chat info for client: %v", err)
	}

	u.livestreamsMu.Lock()
	defer u.livestreamsMu.Unlock()
	u.livestreams[livestreamUrl] = client

	return client, nil
}

func (cb *Chatbot) Run(rule *Rule, url string) error {
	if rule == nil ||
		rule.ChatbotID == nil ||
		rule.ID == nil ||
		rule.Parameters == nil ||
		rule.Parameters.SendAs == nil {
		return pkgErr("", fmt.Errorf("invalid rule"))
	}

	stopped := cb.stopRunner(*rule.ChatbotID, *rule.ID)
	if stopped {
		// TODO: figure out better way to determine when running rule is cleaned up.
		// If rule was stopped, wait for everything to complete before running again.
		time.Sleep(1 * time.Second)
	}

	var err error
	client := cb.clients.byUsernameLivestream(rule.Parameters.SendAs.Username, url)
	if client == nil {
		client, err = cb.addClient(rule.Parameters.SendAs.Username, url)
		if err != nil {
			return pkgErr("error adding client", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := &Runner{
		cancel: cancel,
		client: client,
		rule:   *rule,
		wails:  cb.wails,
	}

	err = cb.initRunner(runner)
	if err != nil {
		return pkgErr("error initializing runner", err)
	}

	go cb.run(ctx, runner)

	return nil
}

func (cb *Chatbot) initRunner(runner *Runner) error {
	if runner == nil || runner.rule.ID == nil || runner.rule.ChatbotID == nil || runner.rule.Parameters == nil || runner.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid runner")
	}

	channelID, err := runner.rule.Parameters.SendAs.ChannelIDInt()
	if err != nil {
		return fmt.Errorf("error converting channel ID to int: %v", err)
	}

	runner.channelIDMu.Lock()
	runner.channelID = channelID
	runner.channelIDMu.Unlock()

	switch {
	case runner.rule.Parameters.Trigger.OnTimer != nil:
		runner.run = runner.runOnTimer
	case runner.rule.Parameters.Trigger.OnCommand != nil:
		err = cb.initRunnerCommand(runner)
		if err != nil {
			return fmt.Errorf("error initializing command: %v", err)
		}
	}

	// cb.runnersMu.Lock()
	// defer cb.runnersMu.Unlock()
	// cb.runners[*runner.rule.ID] = runner

	cb.botsMu.Lock()
	defer cb.botsMu.Unlock()
	bot, exists := cb.bots[*runner.rule.ChatbotID]
	if !exists {
		bot = &Bot{
			runners: map[int64]*Runner{},
		}

		cb.bots[*runner.rule.ChatbotID] = bot
	}

	bot.runnersMu.Lock()
	defer bot.runnersMu.Unlock()
	bot.runners[*runner.rule.ID] = runner

	return nil
}

func (cb *Chatbot) initRunnerCommand(runner *Runner) error {
	runner.run = runner.runOnCommand

	cmd := runner.rule.Parameters.Trigger.OnCommand.Command
	if cmd == "" || cmd[0] != '!' {
		return fmt.Errorf("invalid command")
	}

	chatCh := make(chan events.Chat, 10)
	runner.chatCh = chatCh

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()
	rcvr, exists := cb.receivers[runner.client.LiveStreamUrl]
	if !exists {
		rcvr = &receiver{
			onCommand: map[string]map[int64]chan events.Chat{},
		}
		cb.receivers[runner.client.LiveStreamUrl] = rcvr
	}

	chans, exists := rcvr.onCommand[cmd]
	if !exists {
		chans = map[int64]chan events.Chat{}
		rcvr.onCommand[cmd] = chans
	}
	chans[*runner.rule.ID] = chatCh

	return nil
}

func (cb *Chatbot) run(ctx context.Context, runner *Runner) {
	if runner == nil || runner.rule.ID == nil || runner.run == nil {
		cb.logError.Println("invalid runner")
		return
	}

	runtime.EventsEmit(cb.wails, fmt.Sprintf("ChatbotRuleActive-%d", *runner.rule.ID), true)
	err := runner.run(ctx)
	if err != nil {
		prefix := fmt.Sprintf("chatbot runner for rule %d returned error:", *runner.rule.ID)
		cb.logError.Println(prefix, err)
		runtime.EventsEmit(cb.wails, fmt.Sprintf("ChatbotRuleError-%d", *runner.rule.ID), "Chatbot encountered an error while running this rule.")
	}

	err = cb.stop(&runner.rule)
	if err != nil {
		prefix := fmt.Sprintf("error stopping rule %d after runner returns:", *runner.rule.ID)
		cb.logError.Println(prefix, err)
		return
	}

	runtime.EventsEmit(cb.wails, fmt.Sprintf("ChatbotRuleActive-%d", *runner.rule.ID), false)
}

func (cb *Chatbot) Running(chatbotID int64, ruleID int64) bool {
	// cb.runnersMu.Lock()
	// defer cb.runnersMu.Unlock()
	// _, exists := cb.runners[id]
	// return exists

	cb.botsMu.Lock()
	defer cb.botsMu.Unlock()
	bot, exists := cb.bots[chatbotID]
	if !exists {
		return false
	}

	bot.runnersMu.Lock()
	defer bot.runnersMu.Unlock()
	_, exists = bot.runners[ruleID]
	return exists
}

func (cb *Chatbot) Stop(rule *Rule) error {
	err := cb.stop(rule)
	if err != nil {
		return pkgErr("", err)
	}

	return nil
}

func (cb *Chatbot) stop(rule *Rule) error {
	if rule == nil || rule.ID == nil || rule.ChatbotID == nil {
		return fmt.Errorf("invalid rule")
	}

	cb.stopRunner(*rule.ChatbotID, *rule.ID)

	return nil
}

func (cb *Chatbot) stopRunner(chatbotID int64, ruleID int64) bool {
	// cb.runnersMu.Lock()
	// defer cb.runnersMu.Unlock()
	// runner, exists := cb.runners[id]
	// if !exists {
	// 	return
	// }
	cb.botsMu.Lock()
	defer cb.botsMu.Unlock()
	bot, exists := cb.bots[chatbotID]
	if !exists {
		return false
	}

	bot.runnersMu.Lock()
	defer bot.runnersMu.Unlock()
	runner, exists := bot.runners[ruleID]
	if !exists {
		return false
	}

	stopped := true
	runner.stop()
	// delete(cb.runners, id)
	delete(bot.runners, ruleID)

	switch {
	case runner.rule.Parameters.Trigger.OnCommand != nil:
		err := cb.closeRunnerCommand(runner)
		if err != nil {
			cb.logError.Println("error closing runner command:", err)
		}
	}

	return stopped
}

func (cb *Chatbot) closeRunnerCommand(runner *Runner) error {
	if runner == nil || runner.rule.ID == nil || runner.rule.Parameters == nil || runner.rule.Parameters.Trigger == nil || runner.rule.Parameters.Trigger.OnCommand == nil {
		return fmt.Errorf("invalid runner command")
	}

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	rcvr, exists := cb.receivers[runner.client.LiveStreamUrl]
	if !exists {
		return fmt.Errorf("receiver for runner does not exist")
	}

	cmd := runner.rule.Parameters.Trigger.OnCommand.Command
	chans, exists := rcvr.onCommand[cmd]
	if !exists {
		return fmt.Errorf("channel map for runner does not exist")
	}

	ch, exists := chans[*runner.rule.ID]
	if !exists {
		return fmt.Errorf("channel for runner does not exist")
	}

	close(ch)
	delete(chans, *runner.rule.ID)

	return nil
}

func (cb *Chatbot) HandleChat(event events.Chat) {

	switch event.Message.Type {
	case rumblelivestreamlib.ChatTypeMessages:
		cb.handleMessage(event)
	}
}

func (cb *Chatbot) handleMessage(event events.Chat) {
	errs := cb.runMessageFuncs(
		event,
		cb.handleMessageCommand,
	)

	for _, err := range errs {
		cb.logError.Println("chatbot: error handling message:", err)
	}
}

func (cb *Chatbot) runMessageFuncs(event events.Chat, fns ...messageFunc) []error {
	// TODO: validate message

	errs := []error{}
	for _, fn := range fns {
		err := fn(event)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

type messageFunc func(event events.Chat) error

func (cb *Chatbot) handleMessageCommand(event events.Chat) error {
	if strings.Index(event.Message.Text, "!") != 0 {
		return nil
	}

	words := strings.Split(event.Message.Text, " ")
	cmd := words[0]

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	receiver, exists := cb.receivers[event.Livestream]
	if !exists {
		return nil
	}
	if receiver == nil {
		return fmt.Errorf("receiver is nil for livestream: %s", event.Livestream)
	}

	receiver.onCommandMu.Lock()
	defer receiver.onCommandMu.Unlock()
	runners, exist := receiver.onCommand[cmd]
	if !exist {
		return nil
	}

	for _, runner := range runners {
		runner <- event
	}

	return nil
}
