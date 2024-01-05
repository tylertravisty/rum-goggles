package chatbot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/internal/config"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ChatBot struct {
	ctx        context.Context
	client     *rumblelivestreamlib.Client
	Cfg        config.ChatBot
	logError   *log.Logger
	messages   map[string]*message
	messagesMu sync.Mutex
}

type message struct {
	cancel    context.CancelFunc
	cancelMu  sync.Mutex
	asChannel bool
	id        string
	interval  time.Duration
	text      string
}

func NewChatBot(ctx context.Context, streamUrl string, cfg config.ChatBot, logError *log.Logger) (*ChatBot, error) {
	client, err := rumblelivestreamlib.NewClient("", streamUrl)
	if err != nil {
		return nil, fmt.Errorf("chatbot: error creating new client: %v", err)
	}

	return &ChatBot{ctx: ctx, client: client, Cfg: cfg, logError: logError, messages: map[string]*message{}}, nil
}

func (cb *ChatBot) StartMessage(id string) error {
	msg, exists := cb.Cfg.Messages[id]
	if !exists {
		return fmt.Errorf("chatbot: message does not exist")
	}

	cb.messagesMu.Lock()
	defer cb.messagesMu.Unlock()
	m, exists := cb.messages[id]
	if exists {
		m.stop()
		delete(cb.messages, id)
	}

	m = &message{
		asChannel: msg.AsChannel,
		id:        msg.ID,
		interval:  msg.Interval,
		text:      msg.Text,
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelMu.Lock()
	m.cancel = cancel
	m.cancelMu.Unlock()
	go cb.startMessage(ctx, m)

	cb.messages[id] = m

	return nil
}

func (cb *ChatBot) startMessage(ctx context.Context, m *message) {
	for {
		// TODO: if error, emit error to user, stop loop?
		err := cb.chat(m)
		if err != nil {
			cb.logError.Println("error sending chat:", err)
			cb.StopMessage(m.id)
			runtime.EventsEmit(cb.ctx, "ChatBotMessageError-"+m.id, m.id)
			// TODO: stop this loop?
		} else {
			runtime.EventsEmit(cb.ctx, "ChatBotMessageActive-"+m.id, m.id)
		}

		timer := time.NewTimer(m.interval * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (cb *ChatBot) chat(m *message) error {
	if cb.client == nil {
		return fmt.Errorf("client is nil")
	}

	err := cb.client.Chat(m.asChannel, m.text)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (cb *ChatBot) StopAllMessages() error {
	cb.messagesMu.Lock()
	defer cb.messagesMu.Unlock()

	for id, m := range cb.messages {
		m.stop()
		delete(cb.messages, id)
	}

	return nil
}

func (cb *ChatBot) StopMessage(id string) error {
	cb.messagesMu.Lock()
	defer cb.messagesMu.Unlock()
	m, exists := cb.messages[id]
	if exists {
		fmt.Println("IT EXISTS!!")
		m.stop()
		delete(cb.messages, id)
	}

	return nil
}

func (m *message) stop() {
	m.cancelMu.Lock()
	if m.cancel != nil {
		m.cancel()
	}
	m.cancelMu.Unlock()
}

func (cb *ChatBot) Login(username string, password string) error {
	if cb.client == nil {
		return fmt.Errorf("chatbot: client is nil")
	}

	err := cb.client.Login(username, password)
	if err != nil {
		return fmt.Errorf("chatbot: error logging in: %v", err)
	}

	return nil
}

func (cb *ChatBot) Logout() error {
	if cb.client == nil {
		return fmt.Errorf("chatbot: client is nil")
	}

	err := cb.client.Logout()
	if err != nil {
		return fmt.Errorf("chatbot: error logging out: %v", err)
	}

	return nil
}
