package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/v1/internal/api"
	"github.com/tylertravisty/rum-goggles/v1/internal/config"
	"github.com/tylertravisty/rum-goggles/v1/internal/models"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	_ "github.com/mattn/go-sqlite3"
)

const (
	AccountType = "Account"
	ChannelType = "Channel"
)

// App struct
type App struct {
	api       *api.Api
	clients   map[string]*rumblelivestreamlib.Client
	clientsMu sync.Mutex
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

	app.api = api.NewApi(app.logError, app.logInfo)

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
	a.api.Startup(ctx)

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
	if a.api != nil {
		err := a.api.Shutdown()
		if err != nil {
			a.logError.Println("error shutting down api:", err)
		}
	}

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

func (a *App) AddPage(apiKey string) error {
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
		_, err = a.services.AccountS.Create(&models.Account{
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
	a.clientsMu.Lock()
	defer a.clientsMu.Unlock()
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

	acct, err := a.services.AccountS.ByUsername(username)
	if err != nil {
		a.logError.Println("error getting account by username:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}
	if acct == nil {
		acct = &models.Account{nil, nil, &username, &cookiesS, nil, nil}
		id, err := a.services.AccountS.Create(acct)
		if err != nil {
			a.logError.Println("error creating account:", err)
			return fmt.Errorf("Error logging in. Try again.")
		}
		acct.ID = &id
	} else {
		acct.Cookies = &cookiesS
		err = a.services.AccountS.Update(acct)
		if err != nil {
			a.logError.Println("error updating account:", err)
			return fmt.Errorf("Error logging in. Try again.")
		}
	}

	name := acct.String()
	if name == nil {
		a.logError.Println("account name is nil")
		return fmt.Errorf("Error logging in. Try again.")
	}
	runtime.EventsEmit(a.ctx, "LoggedIn-"+*name, true)

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}
	runtime.EventsEmit(a.ctx, "PageSideBarAccounts", list)

	err = a.openDetails(acct)
	if err != nil {
		a.logError.Println("error opening account details:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}

	return nil
}

func (a *App) Logout(id int64) error {
	acct, err := a.services.AccountS.ByID(id)
	if err != nil {
		a.logError.Println("error querying account by ID:", err)
		return fmt.Errorf("Error logging out. Try again.")
	}
	if acct == nil {
		return fmt.Errorf("Did not find account. Try again.")
	}

	if acct.Username == nil {
		a.logError.Println("account username is nil")
		return fmt.Errorf("Error logging out. Try again.")
	}

	a.clientsMu.Lock()
	defer a.clientsMu.Unlock()
	client, exists := a.clients[*acct.Username]
	if exists {
		err = client.Logout()
		if err != nil {
			a.logError.Println("error logging out:", err)
			return fmt.Errorf("Error logging out. Try again.")
		}
		delete(a.clients, *acct.Username)
	}

	if acct.Cookies != nil {
		acct.Cookies = nil
		err = a.services.AccountS.Update(acct)
		if err != nil {
			a.logError.Println("error updating account:", err)
			return fmt.Errorf("Error logging out. Try again.")
		}
	}

	name := acct.String()
	if name == nil {
		a.logError.Println("account name is nil")
		return fmt.Errorf("Error logging out. Try again.")
	}
	runtime.EventsEmit(a.ctx, "LoggedIn-"+*name, false)

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error logging out. Try again.")
	}
	runtime.EventsEmit(a.ctx, "PageSideBarAccounts", list)

	err = a.openDetails(acct)
	if err != nil {
		a.logError.Println("error opening account details:", err)
		return fmt.Errorf("Error logging out. Try again.")
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
	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return nil, fmt.Errorf("Error retrieving accounts and channels. Try restarting.")
	}

	return list, nil
}

func (a *App) accountList() (map[string]*Account, error) {
	list := map[string]*Account{}

	accountChannels, err := a.services.AccountChannelS.All()
	if err != nil {
		return nil, fmt.Errorf("error querying all account channels: %v", err)
	}

	for _, ac := range accountChannels {
		if ac.Account.Username == nil {
			return nil, fmt.Errorf("account-channel contains nil account username")
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

func (a *App) OpenAccount(id int64) error {
	acct, err := a.services.AccountS.ByID(id)
	if err != nil {
		a.logError.Println("error querying account by ID:", err)
		return fmt.Errorf("Error opening account. Try again.")
	}
	if acct == nil {
		return fmt.Errorf("Did not find account. Try again.")
	}

	err = a.openDetails(acct)
	if err != nil {
		a.logError.Println("error opening account details:", err)
		return fmt.Errorf("Error opening account. Try again.")
	}

	return nil
}

func (a *App) OpenChannel(id int64) error {
	channel, err := a.services.ChannelS.ByID(id)
	if err != nil {
		a.logError.Println("error querying channel by ID:", err)
		return fmt.Errorf("Error opening channel. Try again.")
	}
	if channel == nil {
		return fmt.Errorf("Did not find channel. Try again.")
	}

	err = a.openDetails(channel)
	if err != nil {
		a.logError.Println("error opening channel details:", err)
		return fmt.Errorf("Error opening channel. Try again.")
	}

	return nil
}

type Page interface {
	Id() *int64
	KeyUrl() *string
	LoggedIn() bool
	String() *string
	Title() *string
	Type() string
}

type PageDetails struct {
	ID       int64  `json:"id"`
	HasApi   bool   `json:"has_api"`
	LoggedIn bool   `json:"logged_in"`
	Title    string `json:"title"`
	Type     string `json:"type"`
}

func (a *App) openDetails(p Page) error {
	id := p.Id()
	if id == nil {
		return fmt.Errorf("page id is nil")
	}

	hasApi := true
	key := p.KeyUrl()
	if key == nil || *key == "" {
		hasApi = false
	}

	name := p.String()
	if name == nil {
		return fmt.Errorf("page name is nil")
	}

	title := p.Title()
	if title == nil {
		return fmt.Errorf("page title is nil")
	}

	runtime.EventsEmit(a.ctx, "PageDetails", PageDetails{
		ID:       *id,
		HasApi:   hasApi,
		LoggedIn: p.LoggedIn(),
		Title:    *title,
		Type:     p.Type(),
	})

	err := a.api.Display(*name)
	if err != nil {
		return fmt.Errorf("error displaying api for %s: %v", *name, err)
	}

	return nil
}

func (a *App) ActivateAccount(id int64) error {
	acct, err := a.services.AccountS.ByID(id)
	if err != nil {
		a.logError.Println("error querying account by ID:", err)
		return fmt.Errorf("Error activating account. Try again.")
	}
	if acct == nil {
		return fmt.Errorf("Did not find account. Try again.")
	}

	err = a.activatePage(acct)
	if err != nil {
		a.logError.Println("error activating account:", err)
		return fmt.Errorf("Error activating account. Try again.")
	}

	return nil
}

// ActivateChannel activates an inactivate page and deactivates an active page.
func (a *App) ActivateChannel(id int64) error {
	channel, err := a.services.ChannelS.ByID(id)
	if err != nil {
		a.logError.Println("error querying channel by ID:", err)
		return fmt.Errorf("Error activating channel. Try again.")
	}
	if channel == nil {
		return fmt.Errorf("Did not find channel. Try again.")
	}

	err = a.activatePage(channel)
	if err != nil {
		a.logError.Println("error activating channel:", err)
		return fmt.Errorf("Error activating channel. Try again.")
	}

	return nil
}

// If page is inactivate, activate.
// If page is active, deactivate.
func (a *App) activatePage(p Page) error {
	name := p.String()
	if name == nil {
		return fmt.Errorf("page name is nil")
	}
	url := p.KeyUrl()
	if url == nil {
		return fmt.Errorf("page key url is nil")
	}

	if a.api.Active(*name) {
		err := a.api.Stop(*name)
		if err != nil {
			return fmt.Errorf("error stopping api: %v", err)
		}
		return nil
	}

	err := a.api.Start(*name, *url, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error starting api: %v", err)
	}
	err = a.api.Display(*name)
	if err != nil {
		return fmt.Errorf("error displaying api: %v", err)
	}

	return nil
}

func (a *App) DeleteAccount(id int64) error {
	acct, err := a.services.AccountS.ByID(id)
	if err != nil {
		a.logError.Println("error querying account by ID:", err)
		return fmt.Errorf("Error deleting account. Try again.")
	}
	if acct == nil {
		return fmt.Errorf("Did not find account. Try again.")
	}

	channels, err := a.services.ChannelS.ByAccount(acct)
	if err != nil {
		a.logError.Println("error querying channels by account:", err)
		return fmt.Errorf("Error deleting account. Try again.")
	}
	if len(channels) != 0 {
		return fmt.Errorf("You must delete all channels associated with the account before it can be deleted.")
	}

	name := acct.String()
	if name == nil {
		a.logError.Println("account name is nil")
		return fmt.Errorf("Error deleting account. Try again.")
	}

	if a.api.Active(*name) {
		err := a.api.Stop(*name)
		if err != nil {
			a.logError.Println("error stopping api:", err)
			return fmt.Errorf("Error deleting account. Try again.")
		}
	}

	err = a.services.AccountS.Delete(acct)
	if err != nil {
		a.logError.Println("error deleting account:", err)
		return fmt.Errorf("Error deleting account. Try again.")
	}

	runtime.EventsEmit(a.ctx, "PageDetails", nil)

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error deleting account. Try again.")
	}

	runtime.EventsEmit(a.ctx, "PageSideBarAccounts", list)

	return nil
}

func (a *App) DeleteChannel(id int64) error {
	channel, err := a.services.ChannelS.ByID(id)
	if err != nil {
		a.logError.Println("error querying channel by ID:", err)
		return fmt.Errorf("Error deleting channel. Try again.")
	}
	if channel == nil {
		return fmt.Errorf("Did not find channel. Try again.")
	}

	name := channel.String()
	if name == nil {
		a.logError.Println("channel name is nil")
		return fmt.Errorf("Error deleting channel. Try again.")
	}

	if a.api.Active(*name) {
		err := a.api.Stop(*name)
		if err != nil {
			a.logError.Println("error stopping api:", err)
			return fmt.Errorf("Error deleting channel. Try again.")
		}
	}

	err = a.services.ChannelS.Delete(channel)
	if err != nil {
		a.logError.Println("error deleting channel:", err)
		return fmt.Errorf("Error deleting channel. Try again.")
	}

	runtime.EventsEmit(a.ctx, "PageDetails", nil)

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error deleting channel. Try again.")
	}

	runtime.EventsEmit(a.ctx, "PageSideBarAccounts", list)

	return nil
}

func (a *App) PageStatus(name string) {
	active := false
	isLive := false

	resp := a.api.Response(name)
	if resp != nil {
		active = true
		isLive = len(resp.Livestreams) > 0
	}

	runtime.EventsEmit(a.ctx, "ApiActive-"+name, active)
	runtime.EventsEmit(a.ctx, "PageLive-"+name, isLive)
}

func (a *App) UpdateAccountApi(id int64, apiKey string) error {
	acct, err := a.services.AccountS.ByID(id)
	if err != nil {
		a.logError.Println("error querying account by ID:", err)
		return fmt.Errorf("Error updating account. Try again.")
	}
	if acct == nil {
		return fmt.Errorf("Did not find account. Try again.")
	}

	name := acct.String()
	if name == nil {
		a.logError.Println("account name is nil")
		return fmt.Errorf("Error updating account. Try again.")
	}

	if a.api.Active(*name) {
		err := a.api.Stop(*name)
		if err != nil {
			a.logError.Println("error stopping api:", err)
			return fmt.Errorf("Error updating account. Try again.")
		}
	}

	client := rumblelivestreamlib.Client{StreamKey: apiKey}
	resp, err := client.Request()
	if err != nil {
		a.logError.Println("error executing api request:", err)
		return fmt.Errorf("Error querying API. Verify key and try again.")
	}

	if resp.ChannelName != "" || resp.Username != *acct.Username {
		return fmt.Errorf("API key does not belong to account. Verify key and try again.")
	}

	acct.ApiKey = &apiKey
	err = a.services.AccountS.Update(acct)
	if err != nil {
		a.logError.Println("error updating account:", err)
		return fmt.Errorf("Error updating account. Try again.")
	}

	return nil
}

func (a *App) UpdateChannelApi(id int64, apiKey string) error {
	channel, err := a.services.ChannelS.ByID(id)
	if err != nil {
		a.logError.Println("error querying channel by ID:", err)
		return fmt.Errorf("Error updating channel. Try again.")
	}
	if channel == nil {
		return fmt.Errorf("Did not find channel. Try again.")
	}

	name := channel.String()
	if name == nil {
		a.logError.Println("channel name is nil")
		return fmt.Errorf("Error updating channel. Try again.")
	}

	if a.api.Active(*name) {
		err := a.api.Stop(*name)
		if err != nil {
			a.logError.Println("error stopping api:", err)
			return fmt.Errorf("Error updating channel. Try again.")
		}
	}

	client := rumblelivestreamlib.Client{StreamKey: apiKey}
	resp, err := client.Request()
	if err != nil {
		a.logError.Println("error executing api request:", err)
		return fmt.Errorf("Error querying API. Verify key and try again.")
	}

	if resp.ChannelName != *channel.Name {
		return fmt.Errorf("API key does not belong to channel. Verify key and try again.")
	}

	*channel.ApiKey = apiKey
	err = a.services.ChannelS.Update(channel)
	if err != nil {
		a.logError.Println("error updating channel:", err)
		return fmt.Errorf("Error updating channel. Try again.")
	}

	return nil
}
