package chatbot

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/v1/internal/events"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Runner struct {
	apiCh       chan events.Api
	cancel      context.CancelFunc
	cancelMu    sync.Mutex
	channelID   *int
	channelIDMu sync.Mutex
	chatCh      chan events.Chat
	client      *rumblelivestreamlib.Client
	rule        Rule
	run         runFunc
	wails       context.Context
}

type chatFields struct {
	ChannelName string
	DisplayName string
	Username    string
	Rant        int
}

func (r *Runner) chat(fields *chatFields) error {
	msg, err := r.rule.Parameters.Message.String()
	if err != nil {
		return fmt.Errorf("error getting message string: %v", err)
	}

	if fields != nil {
		tmpl, err := template.New("chat").Parse(msg)
		if err != nil {
			return fmt.Errorf("error creating template: %v", err)
		}

		var msgB bytes.Buffer
		err = tmpl.Execute(&msgB, fields)
		if err != nil {
			return fmt.Errorf("error executing template: %v", err)
		}
		msg = msgB.String()
	}

	err = r.client.Chat(msg, r.channelID)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (r *Runner) init() error {
	if r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid rule")
	}

	channelID, err := r.rule.Parameters.SendAs.ChannelIDInt()
	if err != nil {
		return fmt.Errorf("error converting channel ID to int: %v", err)
	}

	r.channelIDMu.Lock()
	r.channelID = channelID
	r.channelIDMu.Unlock()

	switch {
	case r.rule.Parameters.Trigger.OnTimer != nil:
		r.run = r.runOnTimer
	case r.rule.Parameters.Trigger.OnCommand != nil:
		r.run = r.runOnCommand
	}

	return nil
}

type runFunc func(ctx context.Context) error

func (r *Runner) runOnCommand(ctx context.Context) error {
	if r.rule.ID == nil || r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid rule")
	}
	if r.rule.Parameters.Trigger.OnCommand == nil {
		return fmt.Errorf("command is nil")
	}

	var prev time.Time
	for {
		runtime.EventsEmit(r.wails, fmt.Sprintf("ChatbotRuleActive-%d", *r.rule.ID), true)

		select {
		case <-ctx.Done():
			return nil
		case event := <-r.chatCh:
			now := time.Now()
			if now.Sub(prev) < r.rule.Parameters.Trigger.OnCommand.Timeout*time.Second {
				break
			}

			if block := r.blockCommand(event); block {
				// if bypass := r.bypassCommand(event); !bypass {break}
				break
			}

			err := r.handleCommand(event)
			if err != nil {
				return fmt.Errorf("error handling command: %v", err)
			}
			prev = now
		}
	}
}

func (r *Runner) blockCommand(event events.Chat) bool {
	if r.rule.Parameters.Trigger.OnCommand.Restrict == nil {
		return false
	}

	if r.rule.Parameters.Trigger.OnCommand.Restrict.ToFollower &&
		!event.Message.IsFollower {
		return true
	}

	subscriber := false
	for _, badge := range event.Message.Badges {
		if badge == rumblelivestreamlib.ChatBadgeLocalsSupporter || badge == rumblelivestreamlib.ChatBadgeRecurringSubscription {
			subscriber = true
		}
	}

	if r.rule.Parameters.Trigger.OnCommand.Restrict.ToSubscriber &&
		!subscriber {
		return true
	}

	if event.Message.Rant < r.rule.Parameters.Trigger.OnCommand.Restrict.ToRant*100 {
		return true
	}

	return false
}

func (r *Runner) handleCommand(event events.Chat) error {
	displayName := event.Message.Username
	if event.Message.ChannelName != "" {
		displayName = event.Message.ChannelName
	}

	fields := &chatFields{
		ChannelName: event.Message.ChannelName,
		DisplayName: displayName,
		Username:    event.Message.Username,
		Rant:        event.Message.Rant / 100,
	}

	err := r.chat(fields)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (r *Runner) runOnTimer(ctx context.Context) error {
	if r.rule.ID == nil || r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid rule")
	}
	if r.rule.Parameters.Trigger.OnTimer == nil {
		return fmt.Errorf("timer is nil")
	}

	for {
		runtime.EventsEmit(r.wails, fmt.Sprintf("ChatbotRuleActive-%d", *r.rule.ID), true)
		err := r.chat(nil)
		if err != nil {
			return fmt.Errorf("error sending chat: %v", err)
		}

		trigger := time.NewTimer(*r.rule.Parameters.Trigger.OnTimer * time.Second)
		select {
		case <-ctx.Done():
			trigger.Stop()
			return nil
		case <-trigger.C:
		}
	}
}

func (r *Runner) stop() {
	r.cancelMu.Lock()
	if r.cancel != nil {
		r.cancel()
	}
	r.cancelMu.Unlock()
}
