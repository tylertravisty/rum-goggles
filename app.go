package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/tylertravisty/go-utils/random"
	"github.com/tylertravisty/rum-goggles/internal/config"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
)

const (
	configFilepath = "./config.json"
)

// App struct
type App struct {
	ctx   context.Context
	cfg   *config.App
	cfgMu sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	err := a.loadConfig()
	if err != nil {
		// TODO: handle error better on startup
		log.Fatal("error loading config: ", err)
	}
}

func (a *App) loadConfig() error {
	cfg, err := config.Load(configFilepath)
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
	cfg := &config.App{Channels: []config.Channel{}}
	err := cfg.Save(configFilepath)
	if err != nil {
		return fmt.Errorf("error saving new config: %v", err)
	}

	a.cfg = cfg
	return nil
}

func (a *App) Config() *config.App {
	return a.cfg
}

func (a *App) SaveConfig() error {
	err := a.cfg.Save(configFilepath)
	if err != nil {
		// TODO: log error; return user error
		return fmt.Errorf("Error saving config")
	}

	return nil
}

func (a *App) AddChannel(url string) (*config.App, error) {
	client := rumblelivestreamlib.Client{StreamKey: url}
	resp, err := client.Request()
	if err != nil {
		// TODO: log error
		fmt.Println("error requesting api:", err)
		return nil, fmt.Errorf("error querying API")
	}

	name := resp.Username
	if resp.ChannelName != "" {
		name = resp.ChannelName
	}

	channel := config.Channel{ApiUrl: url, Name: name}

	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	a.cfg.Channels = append(a.cfg.Channels, channel)
	err = a.cfg.Save(configFilepath)
	if err != nil {
		// TODO: log error
		fmt.Println("error saving config:", err)
		return nil, fmt.Errorf("error saving new channel")
	}

	return a.cfg, nil
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	random, err := random.String(10)
	if err != nil {
		fmt.Println("random.Alphabetic err:", err)
		return name
	}
	//return fmt.Sprintf("Hello %s, It's show time!", name)
	return fmt.Sprintf("Hello %s, It's show time!", random)
}

// func (a *App) QueryAPI(url string) (*rumblelivestreamlib.Followers, error) {
// 	fmt.Println("QueryAPI")
// 	client := rumblelivestreamlib.Client{StreamKey: url}
// 	resp, err := client.Request()
// 	if err != nil {
// 		// TODO: log error
// 		fmt.Println("client.Request err:", err)
// 		return nil, fmt.Errorf("API request failed")
// 	}

// 	return &resp.Followers, nil
// }
