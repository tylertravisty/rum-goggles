package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/internal/api"
	"github.com/tylertravisty/rum-goggles/internal/chatbot"
	"github.com/tylertravisty/rum-goggles/internal/config"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type chat struct {
	username string
	password string
	url      string
}

// App struct
type App struct {
	ctx      context.Context
	cfg      *config.App
	cfgMu    sync.Mutex
	api      *api.Api
	apiMu    sync.Mutex
	cb       *chatbot.ChatBot
	cbMu     sync.Mutex
	logError *log.Logger
	logInfo  *log.Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{}
	err := app.initLog()
	if err != nil {
		log.Fatal("error initializing log:", err)
	}

	app.api = api.NewApi(app.logError, app.logInfo)

	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.api.Startup(ctx)

	err := a.loadConfig()
	if err != nil {
		a.logError.Fatal("error loading config: ", err)
	}
}

func (a *App) initLog() error {
	f, err := config.LogFile()
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}

	a.logInfo = log.New(f, "[info]", log.LstdFlags|log.Lshortfile)
	a.logError = log.New(f, "[error]", log.LstdFlags|log.Lshortfile)
	return nil
}

func (a *App) loadConfig() error {
	cfg, err := config.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error loading config: %v", err)
		}

		return a.newConfig()
	}

	a.cfg = cfg
	return nil
}

func (a *App) newConfig() error {
	cfg := &config.App{Channels: map[string]config.Channel{}}
	err := cfg.Save()
	if err != nil {
		return fmt.Errorf("error saving new config: %v", err)
	}

	a.cfg = cfg
	return nil
}

func (a *App) Config() *config.App {
	return a.cfg
}

func (a *App) AddChannel(url string) (*config.App, error) {
	client := rumblelivestreamlib.Client{StreamKey: url}
	resp, err := client.Request()
	if err != nil {
		a.logError.Println("error executing api request:", err)
		return nil, fmt.Errorf("Error querying API. Verify key and try again.")
	}

	name := resp.Username
	if resp.ChannelName != "" {
		name = resp.ChannelName
	}

	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	_, err = a.cfg.NewChannel(url, name)
	if err != nil {
		a.logError.Println("error creating new channel:", err)
		return nil, fmt.Errorf("Error creating new channel. Try again.")
	}

	err = a.cfg.Save()
	if err != nil {
		a.logError.Println("error saving config:", err)
		return nil, fmt.Errorf("Error saving channel information. Try again.")
	}

	return a.cfg, nil
}

func (a *App) ChatBotMessages(cid string) (map[string]config.ChatMessage, error) {
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	channel, exists := a.cfg.Channels[cid]
	if !exists {
		a.logError.Println("channel does not exist:", cid)
		return nil, fmt.Errorf("Cannot find channel. Try reloading.")
	}

	return channel.ChatBot.Messages, nil
}

func (a *App) AddChatMessage(cid string, cm config.ChatMessage) (map[string]config.ChatMessage, error) {
	var err error
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	_, err = a.cfg.NewChatMessage(cid, cm)
	if err != nil {
		a.logError.Println("error creating new chat:", err)
		return nil, fmt.Errorf("Error creating new chat message. Try again.")
	}

	err = a.cfg.Save()
	if err != nil {
		a.logError.Println("error saving config:", err)
		return nil, fmt.Errorf("Error saving chat message information. Try again.")
	}

	a.updateChatBotConfig(a.cfg.Channels[cid].ChatBot)

	return a.cfg.Channels[cid].ChatBot.Messages, nil
}

func (a *App) DeleteChatMessage(cid string, cm config.ChatMessage) (map[string]config.ChatMessage, error) {
	a.cbMu.Lock()
	if a.cb != nil {
		err := a.cb.StopMessage(cm.ID)
		if err != nil {
			a.logError.Println("error stopping chat bot message:", err)
			return nil, fmt.Errorf("Error stopping message. Try Again.")
		}
	}
	a.cbMu.Unlock()

	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	err := a.cfg.DeleteChatMessage(cid, cm)
	if err != nil {
		a.logError.Println("error deleting chat message:", err)
		return nil, fmt.Errorf("Error deleting chat message. Try again.")
	}

	err = a.cfg.Save()
	if err != nil {
		a.logError.Println("error saving config:", err)
		return nil, fmt.Errorf("Error saving chat message information. Try again.")
	}

	a.updateChatBotConfig(a.cfg.Channels[cid].ChatBot)

	return a.cfg.Channels[cid].ChatBot.Messages, nil
}

func (a *App) UpdateChatMessage(cid string, cm config.ChatMessage) (map[string]config.ChatMessage, error) {
	var err error
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	_, err = a.cfg.UpdateChatMessage(cid, cm)
	if err != nil {
		a.logError.Println("error updating chat message:", err)
		return nil, fmt.Errorf("Error updating chat message. Try again.")
	}

	err = a.cfg.Save()
	if err != nil {
		a.logError.Println("error saving config:", err)
		return nil, fmt.Errorf("Error saving chat message information. Try again.")
	}

	a.updateChatBotConfig(a.cfg.Channels[cid].ChatBot)

	return a.cfg.Channels[cid].ChatBot.Messages, nil
}

type NewChatBotResponse struct {
	LoggedIn  bool   `json:"logged_in"`
	StreamUrl string `json:"stream_url"`
	Username  string `json:"username"`
}

func (a *App) GetChatBot(cid string) (NewChatBotResponse, error) {
	if a.cb == nil {
		return NewChatBotResponse{}, fmt.Errorf("Chat bot not initalized.")
	}

	loggedIn, err := a.cb.LoggedIn()
	if err != nil {
		a.logError.Println("error checking if chat bot is logged in:", err)
		return NewChatBotResponse{}, fmt.Errorf("Error checking if chat bot is logged in. Try again.")
	}

	return NewChatBotResponse{loggedIn, a.cb.Cfg.Session.Client.StreamUrl, a.cb.Cfg.Session.Username}, nil
}

func (a *App) NewChatBot(cid string) (NewChatBotResponse, error) {
	a.cbMu.Lock()
	defer a.cbMu.Unlock()

	if a.cb != nil {
		err := a.resetChatBot()
		if err != nil {
			a.logError.Println("error resetting chat bot:", err)
			return NewChatBotResponse{}, fmt.Errorf("Error creating chat bot. Try Again.")
		}
	}
	channel, exists := a.cfg.Channels[cid]
	if !exists {
		a.logError.Println("channel does not exist:", cid)
		return NewChatBotResponse{}, fmt.Errorf("Channel does not exist.")
	}

	if channel.ChatBot.Session.Client.StreamUrl == "" {
		return NewChatBotResponse{}, nil
	}

	var err error
	a.cb, err = chatbot.NewChatBot(a.ctx, channel.ChatBot, a.logError)
	if err != nil {
		a.logError.Println("error creating new chat bot:", err)
		return NewChatBotResponse{}, fmt.Errorf("Error creating new chat bot. Try again.")
	}

	loggedIn, err := a.cb.LoggedIn()
	if err != nil {
		a.logError.Println("error checking if chat bot is logged in:", err)
		return NewChatBotResponse{}, fmt.Errorf("Error checking if chat bot is logged in. Try again.")
	}

	if loggedIn {
		err = a.cb.StartChatStream()
		if err != nil {
			a.logError.Println("error starting chat stream:", err)
			return NewChatBotResponse{}, fmt.Errorf("Error connecting to chat. Try again.")
		}
	}

	return NewChatBotResponse{loggedIn, channel.ChatBot.Session.Client.StreamUrl, channel.ChatBot.Session.Username}, nil
}

func (a *App) LoginChatBot(cid string, username string, password string, streamUrl string) error {
	a.cbMu.Lock()
	defer a.cbMu.Unlock()
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()

	if a.cb != nil {
		err := a.resetChatBot()
		if err != nil {
			a.logError.Println("error resetting chat bot:", err)
			return fmt.Errorf("Error creating chat bot. Try Again.")
		}
	}
	channel, exists := a.cfg.Channels[cid]
	if !exists {
		a.logError.Println("channel does not exist:", cid)
		return fmt.Errorf("Channel does not exist.")
	}
	channel.ChatBot.Session.Client.StreamUrl = streamUrl

	var err error
	a.cb, err = chatbot.NewChatBot(a.ctx, channel.ChatBot, a.logError)
	if err != nil {
		a.logError.Println("error creating new chat bot:", err)
		return fmt.Errorf("Error creating new chat bot. Try again.")
	}

	cookies, err := a.cb.Login(username, password)
	if err != nil {
		a.logError.Println("error logging into chat bot:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}

	channel.ChatBot.Session = config.ChatBotSession{
		Client: rumblelivestreamlib.NewClientOptions{
			Cookies:   cookies,
			StreamUrl: streamUrl,
		},
		Username: username,
	}
	a.cfg.Channels[cid] = channel
	err = a.cfg.Save()
	if err != nil {
		a.logError.Println("error saving config:", err)
		return fmt.Errorf("Error saving session information. Try again.")
	}

	a.cb.Cfg.Session = channel.ChatBot.Session

	err = a.cb.StartChatStream()
	if err != nil {
		a.logError.Println("error starting chat stream:", err)
		return fmt.Errorf("Error connecting to chat. Try again.")
	}

	return nil
}

func (a *App) StopAllChatBot(cid string) error {
	if a.cb == nil {
		return fmt.Errorf("Chat bot not initialized.")
	}

	err := a.cb.StopAllMessages()
	if err != nil {
		a.logError.Println("error stopping all chat bot messages:", err)
		return fmt.Errorf("Error stopping messages.")
	}

	return nil
}

func (a *App) StartAllChatBot(cid string) error {
	if a.cb == nil {
		return fmt.Errorf("Chat bot not initialized.")
	}

	err := a.cb.StartAllMessages()
	if err != nil {
		a.logError.Println("error starting all chat bot messages:", err)
		return fmt.Errorf("Error starting messages.")
	}

	return nil
}

func (a *App) UpdateChatBotUrl(cid string, streamUrl string) error {
	a.cbMu.Lock()
	defer a.cbMu.Unlock()
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()

	if a.cb == nil {
		return fmt.Errorf("Chat bot not initialized.")
	}

	err := a.resetChatBot()
	if err != nil {
		a.logError.Println("error resetting chat bot:", err)
		return fmt.Errorf("Error creating chat bot. Try Again.")
	}

	channel, exists := a.cfg.Channels[cid]
	if !exists {
		a.logError.Println("channel does not exist:", cid)
		return fmt.Errorf("Channel does not exist.")
	}
	channel.ChatBot.Session.Client.StreamUrl = streamUrl

	a.cb, err = chatbot.NewChatBot(a.ctx, channel.ChatBot, a.logError)
	if err != nil {
		a.logError.Println("error creating new chat bot:", err)
		return fmt.Errorf("Error creating new chat bot. Try again.")
	}

	a.cfg.Channels[cid] = channel
	err = a.cfg.Save()
	if err != nil {
		a.logError.Println("error saving config:", err)
		return fmt.Errorf("Error saving session information. Try again.")
	}

	a.cb.Cfg.Session.Client.StreamUrl = streamUrl

	err = a.cb.StartChatStream()
	if err != nil {
		a.logError.Println("error starting chat stream:", err)
		return fmt.Errorf("Error connecting to chat. Try again.")
	}

	return nil
}

func (a *App) ResetChatBot(cid string, logout bool) error {
	a.cbMu.Lock()
	defer a.cbMu.Unlock()
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()

	if a.cb == nil {
		return nil
	}

	err := a.cb.StopAllMessages()
	if err != nil {
		return fmt.Errorf("error stopping all chat bot messages: %v", err)
	}

	if logout {
		err := a.cb.Logout()
		if err != nil {
			return fmt.Errorf("error logging out of chat bot: %v", err)
		}

		//TODO: reset session in config
		channel, exists := a.cfg.Channels[cid]
		if !exists {
			a.logError.Println("channel does not exist:", cid)
			return fmt.Errorf("Channel does not exist.")
		}

		channel.ChatBot.Session = config.ChatBotSession{}
		a.cfg.Channels[cid] = channel
		err = a.cfg.Save()
		if err != nil {
			a.logError.Println("error saving config:", err)
			return fmt.Errorf("Error saving session information. Try again.")
		}
	}

	err = a.resetChatBot()
	if err != nil {
		a.logError.Println("error resetting chat bot:", err)
		return fmt.Errorf("Error resetting chat bot. Try Again.")
	}

	return nil
}

func (a *App) resetChatBot() error {
	if a.cb == nil {
		// return fmt.Errorf("chat bot is nil")
		return nil
	}

	err := a.cb.StopAllMessages()
	if err != nil {
		return fmt.Errorf("error stopping all chat bot messages: %v", err)
	}

	err = a.cb.StopChatStream()
	if err != nil {
		return fmt.Errorf("error stopping chat stream: %v", err)
	}

	a.cb = nil

	return nil
}

func (a *App) StartChatBotMessage(mid string) error {
	a.cbMu.Lock()
	defer a.cbMu.Unlock()

	if a.cb == nil {
		return fmt.Errorf("Chat bot not initialized.")
	}

	err := a.cb.StartMessage(mid)
	if err != nil {
		a.logError.Println("error starting chat bot message:", err)
		return fmt.Errorf("Error starting message. Try Again.")
	}

	return nil
}

func (a *App) StopChatBotMessage(mid string) error {
	a.cbMu.Lock()
	defer a.cbMu.Unlock()

	// If chat bot not initialized, then stop does nothing
	if a.cb == nil {
		return nil
	}

	err := a.cb.StopMessage(mid)
	if err != nil {
		a.logError.Println("error stopping chat bot message:", err)
		return fmt.Errorf("Error stopping message. Try Again.")
	}

	return nil
}

func (a *App) StartApi(cid string) error {
	channel, found := a.cfg.Channels[cid]
	if !found {
		a.logError.Println("could not find channel CID:", cid)
		return fmt.Errorf("channel CID not found")
	}

	err := a.api.Start(channel.ApiUrl, channel.Interval*time.Second)
	if err != nil {
		a.logError.Println("error starting api:", err)
		return fmt.Errorf("error starting API")
	}

	return nil
}

func (a *App) StopApi() {
	a.api.Stop()
}

func (a *App) updateChatBotConfig(cfg config.ChatBot) {
	a.cbMu.Lock()
	defer a.cbMu.Unlock()
	if a.cb != nil {
		a.cb.Cfg = cfg
	}
}

func (a *App) OpenFileDialog() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		a.logError.Println("error getting home directory:", err)
		return "", fmt.Errorf("Error opening file explorer. Try again.")
	}
	filepath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{DefaultDirectory: home})
	if err != nil {
		a.logError.Println("error opening file dialog:", err)
		return "", fmt.Errorf("Error opening file explorer. Try again.")
	}

	return filepath, err
}

func (a *App) FilepathBase(path string) string {
	return filepath.Base(path)
}
