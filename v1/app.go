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
		models.WithChannelService(),
		models.WithAccountChannelService(),
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

func (a *App) AddChannel(apiKey string) error {
	client := rumblelivestreamlib.Client{StreamKey: apiKey}
	resp, err := client.Request()
	if err != nil {
		a.logError.Println("error executing api request:", err)
		return fmt.Errorf("Error querying API. Verify key and try again.")
	}

	userKey := apiKey
	channelKey := ""
	if resp.Type == "channel" {
		userKey = ""
		channelKey = apiKey
	}

	err = a.addAccountNotExist(resp.UserID, resp.Username, userKey)
	if err != nil {
		a.logError.Println("error adding account if not exist:", err)
		return fmt.Errorf("Error adding channel. Try again.")
	}

	if resp.Type == "channel" {
		err = a.addChannelNotExist(resp.Username, fmt.Sprint(resp.ChannelID), resp.ChannelName, channelKey)
		if err != nil {
			a.logError.Println("error adding channel if not exist:", err)
			return fmt.Errorf("Error adding channel. Try again.")
		}
	}

	return nil
}

func (a *App) addAccountNotExist(uid string, username string, apiKey string) error {
	acct, err := a.services.AccountS.ByUsername(username)
	if err != nil {
		return fmt.Errorf("error querying account by username: %v", err)
	}
	if acct == nil {
		err = a.services.AccountS.Create(&models.Account{
			UID:      &uid,
			Username: &username,
			ApiKey:   &apiKey,
		})
		if err != nil {
			return fmt.Errorf("error creating account: %v", err)
		}
	}
	return nil
}

func (a *App) addChannelNotExist(username string, cid string, name string, apiKey string) error {
	channel, err := a.services.ChannelS.ByName(name)
	if err != nil {
		return fmt.Errorf("error querying channel by name: %v", err)
	}
	if channel == nil {
		acct, err := a.services.AccountS.ByUsername(username)
		if err != nil {
			return fmt.Errorf("error querying account by username: %v", err)
		}
		if acct == nil {
			return fmt.Errorf("account does not exist with username: %s", username)
		}

		err = a.services.ChannelS.Create(&models.Channel{
			AccountID: acct.ID,
			CID:       &cid,
			Name:      &name,
			ApiKey:    &apiKey,
		})
		if err != nil {
			return fmt.Errorf("error creating channel: %v", err)
		}
	}

	return nil
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
		act = &models.Account{nil, nil, &username, &cookiesS, nil, nil}
		err = a.services.AccountS.Create(act)
		if err != nil {
			a.logError.Println("error creating account:", err)
			return fmt.Errorf("Error logging in. Try again.")
		}
	} else {
		act.Cookies = &cookiesS
		err = a.services.AccountS.Update(act)
		if err != nil {
			a.logError.Println("error updating account:", err)
			return fmt.Errorf("Error logging in. Try again.")
		}
	}

	return nil
}

func (a *App) SignedIn() (bool, error) {
	accounts, err := a.services.AccountS.All()
	if err != nil {
		a.logError.Println("error getting all accounts:", err)
		return false, fmt.Errorf("Error retrieving accounts. Try restarting.")
	}

	return len(accounts) > 0, nil
}

type Account struct {
	Account  models.Account   `json:"account"`
	Channels []models.Channel `json:"channels"`
}

func (a *App) AccountList() (map[string]*Account, error) {
	list := map[string]*Account{}

	accountChannels, err := a.services.AccountChannelS.All()
	if err != nil {
		a.logError.Println("error getting all account channels:", err)
		return nil, fmt.Errorf("Error retrieving accounts and channels. Try restarting.")
	}

	for _, ac := range accountChannels {
		if ac.Account.Username == nil {
			a.logError.Println("account-channel contains nil account username")
			return nil, fmt.Errorf("Error retrieving accounts and channels. Try restarting.")
		}

		act, exists := list[*ac.Account.Username]
		if !exists || act == nil {
			act = &Account{ac.Account, []models.Channel{}}
			list[*ac.Account.Username] = act
		}

		if ac.Channel.AccountID != nil {
			act.Channels = append(act.Channels, ac.Channel)
		}
	}

	return list, nil
}
