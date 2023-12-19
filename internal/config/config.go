package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Channel struct {
	ApiUrl string `json:"api_url"`
	Name   string `json:"name"`
}

type App struct {
	Channels []Channel `json:"channels"`
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
