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
	logError *log.Logger
	logInfo  *log.Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{}
	err := app.initLog()
	if err != nil {
		log.Fatal("error initializing log")
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
	fp, err := config.LogFile()
	if err != nil {
		return fmt.Errorf("error getting filepath for log file")
	}

	f, err := os.OpenFile(fp, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file")
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
