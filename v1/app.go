package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/tylertravisty/rum-goggles/v1/internal/config"
	"github.com/tylertravisty/rum-goggles/v1/internal/models"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"

	_ "github.com/mattn/go-sqlite3"
)

// App struct
type App struct {
	clients   map[string]*rumblelivestreamlib.Client
	ctx       context.Context
	services  *models.Services
	logError  *log.Logger
	logFile   *os.File
	logFileMu sync.Mutex
	logInfo   *log.Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{
		clients: map[string]*rumblelivestreamlib.Client{},
	}
	err := app.log()
	if err != nil {
		log.Fatal("error initializing log: ", err)
	}

	return app
}

func (a *App) log() error {
	a.logFileMu.Lock()
	defer a.logFileMu.Unlock()

	f, err := config.Log()
	if err != nil {
		return fmt.Errorf("error getting log file: %v", err)
	}

	a.logFile = f
	a.logInfo = log.New(f, "[info]", log.LstdFlags|log.Lshortfile)
	a.logError = log.New(f, "[error]", log.LstdFlags|log.Lshortfile)
	return nil
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	db, err := config.Database()
	if err != nil {
		log.Fatal(err)
	}

	services, err := models.NewServices(
		models.WithDatabase(db),
		models.WithAccountService(),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = services.AutoMigrate()
	if err != nil {
		log.Fatal(err)
	}

	a.services = services

	// TODO: check for update - if available, pop up window
}

func (a *App) shutdown(ctx context.Context) {
	if a.services != nil {
		err := a.services.Close()
		if err != nil {
			a.logError.Println("error closing services:", err)
		}
	}

	a.logFileMu.Lock()
	if a.logFile != nil {
		err := a.logFile.Close()
		if err != nil {
			log.Println("error closing log file:", err)
		}
	}
	a.logFileMu.Unlock()
}

func (a *App) Login(username string, password string) error {
	var err error
	client, exists := a.clients[username]
	if exists && client != nil {
		err = client.Logout()
		if err != nil {
			a.logError.Println("error logging out:", err)
			return fmt.Errorf("Error logging in. Try again.")
		}
	} else {
		client, err = rumblelivestreamlib.NewClient(rumblelivestreamlib.NewClientOptions{})
	}

	cookies, err := client.Login(username, password)
	if err != nil {
		a.logError.Println("error logging in:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}
	a.clients[username] = client

	cookiesB, err := json.Marshal(cookies)
	if err != nil {
		a.logError.Println("error marshaling cookies into json:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}
	cookiesS := string(cookiesB)

	act, err := a.services.AccountS.ByUsername(username)
	if err != nil {
		a.logError.Println("error getting account by username:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}
	if act == nil {
		act = &models.Account{nil, &username, &cookiesS}
		err = a.services.AccountS.Create(act)
		if err != nil {
			a.logError.Println("error creating account:", err)
			return fmt.Errorf("Error logging in. Try again.")
		}
	} else {
		act.Cookies = &cookiesS
		err = a.services.AccountS.Update(act)
		if err != nil {
			a.logError.Println("error updating account", err)
			return fmt.Errorf("Error logging in. Try again.")
		}
	}

	return nil
}
