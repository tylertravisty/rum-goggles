package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/internal/api"
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
	api   *api.Api
	apiMu sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{api: api.NewApi()}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.api.Startup(ctx)
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
	cfg := &config.App{Channels: map[string]config.Channel{}}
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

	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	_, err = a.cfg.NewChannel(url, name)
	if err != nil {
		// TODO: log error
		fmt.Println("error creating new channel:", err)
		return nil, fmt.Errorf("error creating new channel")
	}

	err = a.cfg.Save(configFilepath)
	if err != nil {
		// TODO: log error
		fmt.Println("error saving config:", err)
		return nil, fmt.Errorf("error saving new channel")
	}

	return a.cfg, nil
}

func (a *App) StartApi(cid string) error {
	channel, found := a.cfg.Channels[cid]
	if !found {
		// TODO: log error
		fmt.Println("could not find channel CID:", cid)
		return fmt.Errorf("channel CID not found")
	}

	err := a.api.Start(channel.ApiUrl, channel.Interval*time.Second)
	if err != nil {
		// TODO: log error
		fmt.Println("error starting api:", err)
		return fmt.Errorf("error starting API")
	}

	return nil
}

func (a *App) StopApi() {
	a.api.Stop()
}
