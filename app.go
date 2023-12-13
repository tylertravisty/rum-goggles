package main

import (
	"context"
	"fmt"

	"github.com/tylertravisty/go-utils/random"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
