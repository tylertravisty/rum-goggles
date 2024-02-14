package chatbot

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/internal/config"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ChatBot struct {
	ctx                context.Context
	cancelChatStream   context.CancelFunc
	cancelChatStreamMu sync.Mutex
	client             *rumblelivestreamlib.Client
	commands           map[string]chan rumblelivestreamlib.ChatView
	commandsMu         sync.Mutex
	Cfg                config.ChatBot
	logError           *log.Logger
	messages           map[string]*message
	messagesMu         sync.Mutex
}

type message struct {
	cancel              context.CancelFunc
	cancelMu            sync.Mutex
	asChannel           bool
	command             string
	id                  string
	interval            time.Duration
	onCommand           bool
	onCommandFollower   bool
	onCommandRantAmount int
	OnCommandSubscriber bool
	text                string
	textFromFile        []string
}

func NewChatBot(ctx context.Context, cfg config.ChatBot, logError *log.Logger) (*ChatBot, error) {
	// client, err := rumblelivestreamlib.NewClient("", validUrl(streamUrl))
	client, err := rumblelivestreamlib.NewClient(cfg.Session.Client)

	if err != nil {
		return nil, fmt.Errorf("chatbot: error creating new client: %v", err)
	}

	return &ChatBot{ctx: ctx, client: client, Cfg: cfg, commands: map[string]chan rumblelivestreamlib.ChatView{}, logError: logError, messages: map[string]*message{}}, nil
}

func validUrl(url string) string {
	valid := url
	if !strings.HasPrefix(valid, "https://") {
		valid = "https://" + valid
	}

	return valid
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

	textFromFile := []string{}
	if msg.TextFile != "" {
		file, err := os.Open(msg.TextFile)
		if err != nil {
			return fmt.Errorf("chatbot: error opening file with responses: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			textFromFile = append(textFromFile, line)
		}
	}

	m = &message{
		asChannel:           msg.AsChannel,
		command:             msg.Command,
		id:                  msg.ID,
		interval:            msg.Interval,
		onCommand:           msg.OnCommand,
		onCommandFollower:   msg.OnCommandFollower,
		onCommandRantAmount: msg.OnCommandRantAmount,
		OnCommandSubscriber: msg.OnCommandSubscriber,
		text:                msg.Text,
		textFromFile:        textFromFile,
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelMu.Lock()
	m.cancel = cancel
	m.cancelMu.Unlock()
	if msg.OnCommand {
		go cb.startCommand(ctx, m)
	} else {
		go cb.startMessage(ctx, m)
	}

	cb.messages[id] = m

	return nil
}

func (cb *ChatBot) startCommand(ctx context.Context, m *message) {
	cb.commandsMu.Lock()
	ch := make(chan rumblelivestreamlib.ChatView)
	cb.commands[m.command] = ch
	cb.commandsMu.Unlock()

	var prev time.Time
	for {
		runtime.EventsEmit(cb.ctx, "ChatBotCommandActive-"+m.id, m.id)
		// TODO: if error, emit error to user, stop loop?
		select {
		case <-ctx.Done():
			runtime.EventsEmit(cb.ctx, "ChatBotMessageError-"+m.id, m.id)
			return
		case cv := <-ch:
			if m.onCommandFollower && !cv.IsFollower {
				break
			}

			subscriber := false
			for _, badge := range cv.Badges {
				if badge == "recurring_subscription" || badge == "locals_supporter" {
					subscriber = true
				}
			}

			if m.OnCommandSubscriber && !subscriber {
				break
			}

			// if m.onCommandRantAmount > 0 && cv.Rant < m.onCommandRantAmount * 100 {
			// 	break
			// }

			if cv.Rant < m.onCommandRantAmount*100 {
				break
			}

			// TODO: parse !command
			now := time.Now()
			if now.Sub(prev) < m.interval*time.Second {
				break
			}

			err := cb.chatCommand(m, cv)
			if err != nil {
				cb.logError.Println("error sending chat:", err)
				cb.StopMessage(m.id)
				runtime.EventsEmit(cb.ctx, "ChatBotCommandError-"+m.id, m.id)
				return
			} else {
				prev = now
				// runtime.EventsEmit(cb.ctx, "ChatBotCommandActive-"+m.id, m.id)
			}
		}
	}
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
			runtime.EventsEmit(cb.ctx, "ChatBotMessageError-"+m.id, m.id)
			return
		case <-timer.C:
		}
	}
}

func (cb *ChatBot) chatCommand(m *message, cv rumblelivestreamlib.ChatView) error {
	if cb.client == nil {
		return fmt.Errorf("client is nil")
	}

	msgText := m.text
	if len(m.textFromFile) > 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(m.textFromFile))))
		if err != nil {
			return fmt.Errorf("error generating random number: %v", err)
		}

		msgText = m.textFromFile[n.Int64()]
	}

	tmpl, err := template.New("chat").Parse(msgText)
	if err != nil {
		return fmt.Errorf("error creating template: %v", err)
	}

	fields := struct {
		ChannelName string
		Username    string
		Rant        int
	}{
		ChannelName: cv.ChannelName,
		Username:    cv.Username,
		Rant:        cv.Rant / 100,
	}

	var textB bytes.Buffer
	err = tmpl.Execute(&textB, fields)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}
	text := textB.String()

	err = cb.client.Chat(m.asChannel, text)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (cb *ChatBot) chat(m *message) error {
	if cb.client == nil {
		return fmt.Errorf("client is nil")
	}

	text := m.text
	if len(m.textFromFile) > 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(m.textFromFile))))
		if err != nil {
			return fmt.Errorf("error generating random number: %v", err)
		}

		text = m.textFromFile[n.Int64()]
	}

	err := cb.client.Chat(m.asChannel, text)
	if err != nil {
		return fmt.Errorf("error sending chat: %v", err)
	}

	return nil
}

func (cb *ChatBot) StartAllMessages() error {
	for _, msg := range cb.Cfg.Messages {
		err := cb.StartMessage(msg.ID)
		if err != nil {
			return fmt.Errorf("error starting message: %v", err)
		}
	}

	return nil
}

func (cb *ChatBot) StopAllMessages() error {
	cb.messagesMu.Lock()
	defer cb.messagesMu.Unlock()

	for id, m := range cb.messages {
		m.stop()
		delete(cb.messages, id)

		if m.command != "" && m.onCommand {
			cb.commandsMu.Lock()
			ch, exists := cb.commands[m.command]
			if exists {
				close(ch)
				delete(cb.commands, m.command)
			}
			cb.commandsMu.Unlock()
		}
	}

	return nil
}

func (cb *ChatBot) StopMessage(id string) error {
	cb.messagesMu.Lock()
	defer cb.messagesMu.Unlock()

	m, exists := cb.messages[id]
	if exists {
		m.stop()
		delete(cb.messages, id)

		if m.command != "" && m.onCommand {
			cb.commandsMu.Lock()
			defer cb.commandsMu.Unlock()
			ch, exists := cb.commands[m.command]
			if exists {
				close(ch)
				delete(cb.commands, m.command)
			}
		}
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

func (cb *ChatBot) LoggedIn() (bool, error) {
	if cb.client == nil {
		return false, fmt.Errorf("chatbot: client is nil")
	}

	loggedIn, err := cb.client.LoggedIn()
	if err != nil {
		return false, fmt.Errorf("chatbot: error checking if chat bot is logged in: %v", err)
	}

	return loggedIn, nil
}

func (cb *ChatBot) Login(username string, password string) ([]*http.Cookie, error) {
	if cb.client == nil {
		return nil, fmt.Errorf("chatbot: client is nil")
	}

	cookies, err := cb.client.Login(username, password)
	if err != nil {
		return nil, fmt.Errorf("chatbot: error logging in: %v", err)
	}

	return cookies, nil
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

func (cb *ChatBot) StartChatStream() error {
	if cb.client == nil {
		return fmt.Errorf("chatbot: client is nil")
	}

	err := cb.client.ChatInfo()
	if err != nil {
		return fmt.Errorf("chatbot: error getting chat info: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cb.cancelChatStreamMu.Lock()
	cb.cancelChatStream = cancel
	cb.cancelChatStreamMu.Unlock()

	go cb.startChatStream(ctx)

	// err = cb.client.StartChatStream(cb.handleChat, cb.handleError)
	// if err != nil {
	// 	return fmt.Errorf("chatbot: error starting chat stream: %v", err)
	// }

	return nil
}

func (cb *ChatBot) startChatStream(ctx context.Context) {
	for {
		err := cb.client.StartChatStream(cb.handleChat, cb.handleError)
		if err != nil {
			cb.logError.Println("error starting chat stream:", err)
			runtime.EventsEmit(cb.ctx, "ChatBotChatStreamError", "Error starting chat stream.")
			return
		}
		select {
		case <-time.After(90 * time.Minute):
			cb.client.StopChatStream()
			break
		case <-ctx.Done():
			cb.client.StopChatStream()
			return
		}
	}
}

func (cb *ChatBot) StopChatStream() error {
	if cb.client == nil {
		return fmt.Errorf("chatbot: client is nil")
	}

	// TODO: should a panic be caught here?
	cb.cancelChatStreamMu.Lock()
	if cb.cancelChatStream != nil {
		cb.cancelChatStream()
	} else {
		cb.client.StopChatStream()
	}
	cb.cancelChatStreamMu.Unlock()

	return nil
}

func (cb *ChatBot) RestartChatStream() error {
	if cb.client == nil {
		return fmt.Errorf("chatbot: client is nil")
	}

	cb.client.StopChatStream()

	err := cb.client.StartChatStream(cb.handleChat, cb.handleError)
	if err != nil {
		return fmt.Errorf("chatbot: error starting chat stream: %v", err)
	}

	return nil
}

func (cb *ChatBot) handleChat(cv rumblelivestreamlib.ChatView) {
	// runtime.EventsEmit(cb.ctx, "ChatMessageReceived", cv)

	if cv.Type != "init" {
		cb.handleCommand(cv)
	}
}

func (cb *ChatBot) handleCommand(cv rumblelivestreamlib.ChatView) {
	cb.commandsMu.Lock()
	defer cb.commandsMu.Unlock()

	words := strings.Split(cv.Text, " ")
	first := words[0]
	cmd, exists := cb.commands[first]
	if !exists {
		return
	}

	select {
	case cmd <- cv:
		return
	default:
		return
	}
}

func (cb *ChatBot) handleError(err error) {
	cb.logError.Println("error handling chat message:", err)
	// runtime.EventsEmit(cb.ctx, "ChatError", err)
}
