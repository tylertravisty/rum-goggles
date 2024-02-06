package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tylertravisty/go-utils/random"
)

const (
	CIDLen          = 8
	DefaultInterval = 10

	configDir    = ".rum-goggles"
	configDirWin = "RumGoggles"
	configFile   = "config.json"
	logFile      = "logs.txt"
)

func LogFile() (*os.File, error) {
	dir, err := buildConfigDir()
	if err != nil {
		return nil, fmt.Errorf("config: error getting config directory: %v", err)
	}

	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return nil, fmt.Errorf("config: error making config directory: %v", err)
	}

	fp := filepath.Join(dir, logFile)

	f, err := os.OpenFile(fp, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("config: error opening log file: %v", err)
	}

	return f, nil
}

func buildConfigDir() (string, error) {
	userDir, err := userDir()
	if err != nil {
		return "", fmt.Errorf("error getting user directory: %v", err)
	}

	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = filepath.Join(userDir, configDirWin)
	default:
		dir = filepath.Join(userDir, configDir)
	}

	return dir, nil
}

func userDir() (string, error) {
	var dir string
	var err error
	switch runtime.GOOS {
	case "windows":
		dir, err = os.UserCacheDir()
	default:
		dir, err = os.UserHomeDir()
	}

	return dir, err
}

type ChatMessage struct {
	ID                  string        `json:"id"`
	AsChannel           bool          `json:"as_channel"`
	Command             string        `json:"command"`
	Interval            time.Duration `json:"interval"`
	OnCommand           bool          `json:"on_command"`
	OnCommandFollower   bool          `json:"on_command_follower"`
	OnCommandRantAmount int           `json:"on_command_rant_amount"`
	OnCommandSubscriber bool          `json:"on_command_subscriber"`
	Text                string        `json:"text"`
	TextFile            string        `json:"text_file"`
}

type ChatBot struct {
	Messages map[string]ChatMessage `json:"messages"`
	// Commands []ChatCommand
}

type Channel struct {
	ID       string        `json:"id"`
	ApiUrl   string        `json:"api_url"`
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
	ChatBot  ChatBot       `json:"chat_bot"`
}

func (a *App) NewChannel(url string, name string) (string, error) {
	for {
		id, err := random.String(CIDLen)
		if err != nil {
			return "", fmt.Errorf("config: error generating ID: %v", err)
		}

		if _, exists := a.Channels[id]; !exists {
			a.Channels[id] = Channel{id, url, name, DefaultInterval, ChatBot{map[string]ChatMessage{}}}
			return id, nil
		}
	}
}

func (a *App) DeleteChatMessage(cid string, cm ChatMessage) error {
	channel, exists := a.Channels[cid]
	if !exists {
		return fmt.Errorf("config: channel does not exist")
	}

	_, exists = channel.ChatBot.Messages[cm.ID]
	if !exists {
		return fmt.Errorf("config: message does not exist")
	}

	delete(channel.ChatBot.Messages, cm.ID)

	return nil
}

func (a *App) NewChatMessage(cid string, cm ChatMessage) (string, error) {
	if _, exists := a.Channels[cid]; !exists {
		return "", fmt.Errorf("config: channel does not exist")
	}

	for {
		id, err := random.String(CIDLen)
		if err != nil {
			return "", fmt.Errorf("config: error generating ID: %v", err)
		}

		_, exists := a.Channels[cid].ChatBot.Messages[id]
		if !exists {
			cm.ID = id
			a.Channels[cid].ChatBot.Messages[id] = cm
			return id, nil
		}
	}
}

func (a *App) UpdateChatMessage(cid string, cm ChatMessage) (string, error) {
	channel, exists := a.Channels[cid]
	if !exists {
		return "", fmt.Errorf("config: channel does not exist")
	}

	_, exists = channel.ChatBot.Messages[cm.ID]
	if !exists {
		return "", fmt.Errorf("config: message does not exist")
	}

	channel.ChatBot.Messages[cm.ID] = cm

	return cm.ID, nil
}

type App struct {
	Channels map[string]Channel `json:"channels"`
}

func Load() (*App, error) {
	dir, err := buildConfigDir()
	if err != nil {
		return nil, fmt.Errorf("config: error getting config directory: %v", err)
	}

	fp := filepath.Join(dir, configFile)
	app, err := load(fp)
	if err != nil {
		return nil, fmt.Errorf("config: error loading config: %w", err)
	}

	return app, nil
}

func load(filepath string) (*App, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	var app App
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&app)
	if err != nil {
		return nil, fmt.Errorf("error decoding file into json: %v", err)
	}

	return &app, nil
}

func (a *App) Save() error {
	dir, err := buildConfigDir()
	if err != nil {
		return fmt.Errorf("config: error getting config directory: %v", err)
	}

	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return fmt.Errorf("config: error making config directory: %v", err)
	}

	fp := filepath.Join(dir, configFile)
	err = a.save(fp)
	if err != nil {
		return fmt.Errorf("config: error saving config: %v", err)
	}

	return nil
}

func (app *App) save(filepath string) error {
	b, err := json.MarshalIndent(app, "", "\t")
	if err != nil {
		return fmt.Errorf("error encoding config into json: %v", err)
	}

	err = os.WriteFile(filepath, b, 0666)
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}
