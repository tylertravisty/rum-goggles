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
	apiCh       chan events.ApiFollower
	cancel      context.CancelFunc
	cancelMu    sync.Mutex
	channelID   *int
	channelIDMu sync.Mutex
	chatCh      chan events.Chat
	client      *rumblelivestreamlib.Client
	page        string
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

// func (r *Runner) init() error {
// 	if r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
// 		return fmt.Errorf("invalid rule")
// 	}

// 	channelID, err := r.rule.Parameters.SendAs.ChannelIDInt()
// 	if err != nil {
// 		return fmt.Errorf("error converting channel ID to int: %v", err)
// 	}

// 	r.channelIDMu.Lock()
// 	r.channelID = channelID
// 	r.channelIDMu.Unlock()

// 	switch {
// 	case r.rule.Parameters.Trigger.OnTimer != nil:
// 		r.run = r.runOnTimer
// 	case r.rule.Parameters.Trigger.OnEvent != nil:
// 		r.run = r.runOnEvent
// 	case r.rule.Parameters.Trigger.OnCommand != nil:
// 		r.run = r.runOnCommand
// 	}

// 	return nil
// }

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
		case chat := <-r.chatCh:
			now := time.Now()
			if now.Sub(prev) < r.rule.Parameters.Trigger.OnCommand.Timeout*time.Second {
				break
			}

			if block := r.blockCommand(chat); block {
				// if bypass := r.bypassCommand(chat); !bypass {break}
				break
			}

			err := r.handleCommand(chat)
			if err != nil {
				return fmt.Errorf("error handling command: %v", err)
			}
			prev = now
		}
	}
}

func (r *Runner) blockCommand(chat events.Chat) bool {
	if r.rule.Parameters.Trigger.OnCommand.Restrict == nil {
		return false
	}

	if r.rule.Parameters.Trigger.OnCommand.Restrict.ToFollower &&
		!chat.Message.IsFollower {
		return true
	}

	subscriber := false
	for _, badge := range chat.Message.Badges {
		if badge == rumblelivestreamlib.ChatBadgeLocalsSupporter || badge == rumblelivestreamlib.ChatBadgeRecurringSubscription {
			subscriber = true
		}
	}

	if r.rule.Parameters.Trigger.OnCommand.Restrict.ToSubscriber &&
		!subscriber {
		return true
	}

	if chat.Message.Rant < r.rule.Parameters.Trigger.OnCommand.Restrict.ToRant*100 {
		return true
	}

	return false
}

func (r *Runner) handleCommand(chat events.Chat) error {
	displayName := chat.Message.Username
	if chat.Message.ChannelName != "" {
		displayName = chat.Message.ChannelName
	}

	fields := &chatFields{
		ChannelName: chat.Message.ChannelName,
		DisplayName: displayName,
		Username:    chat.Message.Username,
		Rant:        chat.Message.Rant / 100,
	}

	err := r.chat(fields)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (r *Runner) runOnEventFromAccountOnFollow(ctx context.Context) error {
	if r.rule.ID == nil || r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid rule")
	}
	if r.rule.Parameters.Trigger.OnEvent == nil || r.rule.Parameters.Trigger.OnEvent.FromAccount == nil || r.rule.Parameters.Trigger.OnEvent.FromAccount.OnFollow == nil {
		return fmt.Errorf("event is nil")
	}

	for {
		runtime.EventsEmit(r.wails, fmt.Sprintf("ChatbotRuleActive-%d", *r.rule.ID), true)

		select {
		case <-ctx.Done():
			return nil
		case api := <-r.apiCh:
			err := r.handleEventOnFollow(api)
			if err != nil {
				return fmt.Errorf("error handling event: %v", err)
			}
		}
	}
}

func (r *Runner) runOnEventFromChannelOnFollow(ctx context.Context) error {
	if r.rule.ID == nil || r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid rule")
	}
	if r.rule.Parameters.Trigger.OnEvent == nil || r.rule.Parameters.Trigger.OnEvent.FromChannel == nil || r.rule.Parameters.Trigger.OnEvent.FromChannel.OnFollow == nil {
		return fmt.Errorf("event is nil")
	}

	for {
		runtime.EventsEmit(r.wails, fmt.Sprintf("ChatbotRuleActive-%d", *r.rule.ID), true)

		select {
		case <-ctx.Done():
			return nil
		case api := <-r.apiCh:
			err := r.handleEventOnFollow(api)
			if err != nil {
				return fmt.Errorf("error handling event: %v", err)
			}
		}
	}
}

func (r *Runner) handleEventOnFollow(follower events.ApiFollower) error {
	fields := &chatFields{
		DisplayName: follower.Username,
		Username:    follower.Username,
	}

	err := r.chat(fields)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (r *Runner) runOnEventFromLiveStreamOnRaid(ctx context.Context) error {
	if r.rule.ID == nil || r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid rule")
	}
	if r.rule.Parameters.Trigger.OnEvent == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream.OnRaid == nil {
		return fmt.Errorf("event is nil")
	}

	for {
		runtime.EventsEmit(r.wails, fmt.Sprintf("ChatbotRuleActive-%d", *r.rule.ID), true)

		select {
		case <-ctx.Done():
			return nil
		case chat := <-r.chatCh:
			err := r.handleEventFromLiveStreamOnRaid(chat)
			if err != nil {
				return fmt.Errorf("error handling event: %v", err)
			}
		}
	}
}

func (r *Runner) handleEventFromLiveStreamOnRaid(chat events.Chat) error {
	if r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil || r.rule.Parameters.Trigger.OnEvent == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream.OnRaid == nil {
		return fmt.Errorf("invalid rule")
	}

	displayName := chat.Message.Username
	if chat.Message.ChannelName != "" {
		displayName = chat.Message.ChannelName
	}

	fields := &chatFields{
		ChannelName: chat.Message.ChannelName,
		DisplayName: displayName,
		Username:    chat.Message.Username,
		Rant:        chat.Message.Rant / 100,
	}

	err := r.chat(fields)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (r *Runner) runOnEventFromLiveStreamOnRant(ctx context.Context) error {
	if r.rule.ID == nil || r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid rule")
	}
	if r.rule.Parameters.Trigger.OnEvent == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream.OnRant == nil {
		return fmt.Errorf("event is nil")
	}

	for {
		runtime.EventsEmit(r.wails, fmt.Sprintf("ChatbotRuleActive-%d", *r.rule.ID), true)

		select {
		case <-ctx.Done():
			return nil
		case chat := <-r.chatCh:
			err := r.handleEventFromLiveStreamOnRant(chat)
			if err != nil {
				return fmt.Errorf("error handling event: %v", err)
			}
		}
	}
}

func (r *Runner) handleEventFromLiveStreamOnRant(chat events.Chat) error {
	if r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil || r.rule.Parameters.Trigger.OnEvent == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream.OnRant == nil {
		return fmt.Errorf("invalid rule")
	}

	rant := chat.Message.Rant / 100
	minAmount := r.rule.Parameters.Trigger.OnEvent.FromLiveStream.OnRant.MinAmount
	maxAmount := r.rule.Parameters.Trigger.OnEvent.FromLiveStream.OnRant.MaxAmount
	if minAmount != 0 && rant < minAmount {
		return nil
	}
	if maxAmount != 0 && rant > maxAmount {
		return nil
	}

	displayName := chat.Message.Username
	if chat.Message.ChannelName != "" {
		displayName = chat.Message.ChannelName
	}

	fields := &chatFields{
		ChannelName: chat.Message.ChannelName,
		DisplayName: displayName,
		Username:    chat.Message.Username,
		Rant:        chat.Message.Rant / 100,
	}

	err := r.chat(fields)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (r *Runner) runOnEventFromLiveStreamOnSub(ctx context.Context) error {
	if r.rule.ID == nil || r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid rule")
	}
	if r.rule.Parameters.Trigger.OnEvent == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream.OnSub == nil {
		return fmt.Errorf("event is nil")
	}

	for {
		runtime.EventsEmit(r.wails, fmt.Sprintf("ChatbotRuleActive-%d", *r.rule.ID), true)

		select {
		case <-ctx.Done():
			return nil
		case chat := <-r.chatCh:
			err := r.handleEventFromLiveStreamOnSub(chat)
			if err != nil {
				return fmt.Errorf("error handling event: %v", err)
			}
		}
	}
}

func (r *Runner) handleEventFromLiveStreamOnSub(chat events.Chat) error {
	if r.rule.Parameters == nil || r.rule.Parameters.Trigger == nil || r.rule.Parameters.Trigger.OnEvent == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream == nil || r.rule.Parameters.Trigger.OnEvent.FromLiveStream.OnSub == nil {
		return fmt.Errorf("invalid rule")
	}

	displayName := chat.Message.Username
	if chat.Message.ChannelName != "" {
		displayName = chat.Message.ChannelName
	}

	fields := &chatFields{
		ChannelName: chat.Message.ChannelName,
		DisplayName: displayName,
		Username:    chat.Message.Username,
		Rant:        chat.Message.Rant / 100,
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
