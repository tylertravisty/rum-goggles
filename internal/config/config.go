package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tylertravisty/go-utils/random"
)

const (
	CIDLen          = 8
	DefaultInterval = 10
)

type Channel struct {
	ID       string        `json:"id"`
	ApiUrl   string        `json:"api_url"`
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
}

func (a *App) NewChannel(url string, name string) (string, error) {
	for {
		id, err := random.String(CIDLen)
		if err != nil {
			return "", fmt.Errorf("config: error generating ID: %v", err)
		}

		if _, exists := a.Channels[id]; !exists {
			a.Channels[id] = Channel{id, url, name, DefaultInterval}
			return id, nil
		}
	}
}

type App struct {
	Channels map[string]Channel `json:"channels"`
}

func Load(filepath string) (*App, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("config: error opening file: %w", err)
	}

	var app App
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&app)
	if err != nil {
		return nil, fmt.Errorf("config: error decoding file into json: %v", err)
	}

	return &app, nil
}

func (app *App) Save(filepath string) error {
	b, err := json.MarshalIndent(app, "", "\t")
	if err != nil {
		return fmt.Errorf("config: error encoding config into json: %v", err)
	}

	err = os.WriteFile(filepath, b, 0666)
	if err != nil {
		return fmt.Errorf("config: error writing config file: %v", err)
	}

	return nil
}
