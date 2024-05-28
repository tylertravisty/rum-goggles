package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/v1/internal/chatbot"
	"github.com/tylertravisty/rum-goggles/v1/internal/config"
	"github.com/tylertravisty/rum-goggles/v1/internal/events"
	"github.com/tylertravisty/rum-goggles/v1/internal/models"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	_ "github.com/mattn/go-sqlite3"
)

const (
	AccountType = "Account"
	ChannelType = "Channel"
)

type ApiState struct {
	active   bool
	activeMu sync.Mutex
	resp     *rumblelivestreamlib.LivestreamResponse
	respMu   sync.Mutex
}

type Page struct {
	active       bool
	activeMu     sync.Mutex
	apiSt        *ApiState
	displaying   bool
	displayingMu sync.Mutex
	name         string
}

func (p *Page) staticLiveStreamUrl() string {
	return fmt.Sprintf("https://rumble.com%s/live", p.name)
}

// App struct
type App struct {
	cancelProc   context.CancelFunc
	chatbot      *chatbot.Chatbot
	clients      map[string]*rumblelivestreamlib.Client
	clientsMu    sync.Mutex
	displaying   string
	displayingMu sync.Mutex
	logError     *log.Logger
	logFile      *os.File
	logFileMu    sync.Mutex
	logInfo      *log.Logger
	pages        map[string]*Page
	pagesMu      sync.Mutex
	producers    *events.Producers
	services     *models.Services
	wails        context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{
		clients: map[string]*rumblelivestreamlib.Client{},
		pages:   map[string]*Page{},
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
func (a *App) startup(wails context.Context) {
	a.wails = wails
}

func (a *App) process(ctx context.Context) {
	for {
		select {
		case apiE := <-a.producers.ApiP.Ch:
			a.processApi(apiE)
		case chatE := <-a.producers.ChatP.Ch:
			a.processChat(chatE)
		case <-ctx.Done():
			return
		}
	}
}

type apiProcessor func(event events.Api)

func (a *App) runApiProcessors(event events.Api, procs ...apiProcessor) {
	for _, proc := range procs {
		proc(event)
	}
}

func (a *App) processApi(event events.Api) {
	a.runApiProcessors(
		event,
		a.pageApiProcessor,
		a.chatbotApiProcessor,
	)
}

func (a *App) pageApiProcessor(event events.Api) {
	if event.Name == "" {
		a.logError.Println("page cannot process API: event name is empty")
	}

	a.pagesMu.Lock()
	defer a.pagesMu.Unlock()
	page, exists := a.pages[event.Name]
	if !exists {
		page = &Page{
			apiSt: &ApiState{},
			name:  event.Name,
		}
		a.pages[event.Name] = page
	}

	page.apiSt.activeMu.Lock()
	page.apiSt.active = !event.Stop
	page.apiSt.activeMu.Unlock()

	if event.Stop {
		runtime.EventsEmit(a.wails, "ApiActive-"+page.name, false)
		return
	}

	runtime.EventsEmit(a.wails, "ApiActive-"+page.name, true)
	page.displayingMu.Lock()
	if page.displaying {
		runtime.EventsEmit(a.wails, "PageActive", true)
	}
	page.displayingMu.Unlock()

	page.apiSt.respMu.Lock()
	page.apiSt.resp = event.Resp
	page.apiSt.respMu.Unlock()

	a.updatePage(page)
}

type chatProcessor func(event events.Chat)

func (a *App) runChatProcessors(event events.Chat, procs ...chatProcessor) {
	for _, proc := range procs {
		proc(event)
	}
}

func (a *App) processChat(event events.Chat) {
	if event.Stop {
		runtime.EventsEmit(a.wails, "ChatStreamActive-"+event.Url, false)
		return
	}

	a.runChatProcessors(
		event,
		a.chatbotChatProcessor,
	)
}

// TODO: implement this
func (a *App) chatbotApiProcessor(event events.Api) {
	a.chatbot.HandleApi(event)
}

func (a *App) chatbotChatProcessor(event events.Chat) {
	if event.Message.Type == rumblelivestreamlib.ChatTypeInit {
		return
	}

	a.chatbot.HandleChat(event)
}

func (a *App) shutdown(ctx context.Context) {
	err := a.producers.Shutdown()
	if err != nil {
		a.logError.Println("error closing event producers:", err)
	}

	a.cancelProc()

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

func (a *App) Start() (bool, error) {

	runtime.EventsEmit(a.wails, "StartupMessage", "Initializing database...")
	err := a.initServices()
	if err != nil {
		a.logError.Println("error initializing services:", err)
		return false, fmt.Errorf("Error starting Rum Goggles. Try restarting.")
	}
	runtime.EventsEmit(a.wails, "StartupMessage", "Initializing database complete.")

	runtime.EventsEmit(a.wails, "StartupMessage", "Verifying account sessions...")
	count, err := a.verifyAccounts()
	if err != nil {
		a.logError.Println("error verifying accounts:", err)
		return false, fmt.Errorf("Error starting Rum Goggles. Try restarting.")
	}
	runtime.EventsEmit(a.wails, "StartupMessage", "Verifying account sessions complete.")

	runtime.EventsEmit(a.wails, "StartupMessage", "Initializing event producers...")
	err = a.initProducers()
	if err != nil {
		a.logError.Println("error initializing producers:", err)
		return false, fmt.Errorf("Error starting Rum Goggles. Try restarting.")
	}
	runtime.EventsEmit(a.wails, "StartupMessage", "Initializing event producers complete.")

	runtime.EventsEmit(a.wails, "StartupMessage", "Initializing chat bot...")
	err = a.initChatbot()
	if err != nil {
		a.logError.Println("error initializing chat bot:", err)
		return false, fmt.Errorf("Error starting Rum Goggles. Try restarting.")
	}
	runtime.EventsEmit(a.wails, "StartupMessage", "Initializing chat bot complete.")

	// TODO: check for update - if available, pop up window
	// runtime.EventsEmit(a.ctx, "StartupMessage", "Checking for updates...")
	// update, err = a.checkForUpdate()
	// runtime.EventsEmit(a.ctx, "StartupMessage", "Checking for updates complete.")

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelProc = cancel
	go a.process(ctx)

	signin := true
	if count > 0 {
		signin = false
	}

	return signin, nil
}

func (a *App) initChatbot() error {
	cb := chatbot.New(a.services.AccountS, a.services.ChatbotS, a.logError, a.wails)
	a.chatbot = cb

	return nil
}

func (a *App) initProducers() error {
	producers, err := events.NewProducers(
		events.WithLoggers(a.logError, a.logInfo),
		events.WithApiProducer(),
		events.WithChatProducer(),
	)
	if err != nil {
		return fmt.Errorf("error initializing producers: %v", err)
	}

	err = producers.Startup()
	if err != nil {
		return fmt.Errorf("error starting producers: %v", err)
	}

	a.producers = producers

	return nil
}

func (a *App) initServices() error {
	db, err := config.Database()
	if err != nil {
		return fmt.Errorf("error getting database config: %v", err)
	}

	services, err := models.NewServices(
		models.WithDatabase(db),
		models.WithAccountService(),
		models.WithChannelService(),
		models.WithAccountChannelService(),
		models.WithChatbotService(),
		models.WithChatbotRuleService(),
	)
	if err != nil {
		return fmt.Errorf("error initializing services: %v", err)
	}

	err = services.AutoMigrate()
	if err != nil {
		return fmt.Errorf("error auto-migrating services: %v", err)
	}

	a.services = services

	return nil
}

func (a *App) verifyAccounts() (int, error) {
	accounts, err := a.services.AccountS.All()
	if err != nil {
		return -1, fmt.Errorf("error querying all accounts: %v", err)
	}

	a.clientsMu.Lock()
	defer a.clientsMu.Unlock()
	for _, account := range accounts {
		if account.Cookies != nil {
			var cookies []*http.Cookie
			err = json.Unmarshal([]byte(*account.Cookies), &cookies)
			if err != nil {
				return -1, fmt.Errorf("error un-marshaling cookie string: %v", err)
			}
			client, err := rumblelivestreamlib.NewClient(rumblelivestreamlib.NewClientOptions{Cookies: cookies})
			if err != nil {
				return -1, fmt.Errorf("error creating new client: %v", err)
			}
			if account.Username == nil {
				return -1, fmt.Errorf("account username is nil")
			}
			loggedIn, err := client.LoggedIn()
			if err != nil {
				return -1, fmt.Errorf("error checking if account is logged in: %v", err)
			}
			if loggedIn.User.LoggedIn {
				a.clients[*account.Username] = client
			} else {
				account.Cookies = nil
				err = a.services.AccountS.Update(&account)
				if err != nil {
					return -1, fmt.Errorf("error updating account: %v", err)
				}
			}
		}
	}

	return len(accounts), nil
}

func (a *App) AddPage(apiKey string) error {
	client := rumblelivestreamlib.Client{ApiKey: apiKey}
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

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}
	runtime.EventsEmit(a.wails, "PageSideBarAccounts", list)

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
		loggedIn, err := client.LoggedIn()
		if err != nil {
			a.logError.Println("error checking if account is logged in:", err)
			return fmt.Errorf("Error logging in. Try again.")
		}

		uid, found := strings.CutPrefix(loggedIn.User.ID, "_u")
		if !found {
			a.logError.Println("did not find uid prefix '_u' in response after checking if accounts is logged in")
			return fmt.Errorf("Error logging in. Try again.")
		}
		rumbleUsername := loggedIn.Data.Username
		if rumbleUsername == "" {
			a.logError.Println("username is empty in response after checking if accounts is logged in")
			return fmt.Errorf("Error logging in. Try again.")
		}

		acct = &models.Account{nil, &uid, &rumbleUsername, &cookiesS, nil, nil}
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
	runtime.EventsEmit(a.wails, "LoggedIn-"+*name, true)

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error logging in. Try again.")
	}
	runtime.EventsEmit(a.wails, "PageSideBarAccounts", list)

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
		if !exists {
			var cookies []*http.Cookie
			err = json.Unmarshal([]byte(*acct.Cookies), &cookies)
			if err != nil {
				a.logError.Println("error un-marshaling cookie string:", err)
				return fmt.Errorf("Error logging out. Try again.")
			}
			client, err = rumblelivestreamlib.NewClient(rumblelivestreamlib.NewClientOptions{Cookies: cookies})
			err = client.Logout()
			if err != nil {
				a.logError.Println("error logging out:", err)
				return fmt.Errorf("Error logging out. Try again.")
			}
		}
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
	runtime.EventsEmit(a.wails, "LoggedIn-"+*name, false)

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error logging out. Try again.")
	}
	runtime.EventsEmit(a.wails, "PageSideBarAccounts", list)

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

type PageInfo interface {
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

func (a *App) openDetails(pi PageInfo) error {
	id := pi.Id()
	if id == nil {
		return fmt.Errorf("page id is nil")
	}

	hasApi := true
	key := pi.KeyUrl()
	if key == nil || *key == "" {
		hasApi = false
	}

	name := pi.String()
	if name == nil {
		return fmt.Errorf("page name is nil")
	}

	title := pi.Title()
	if title == nil {
		return fmt.Errorf("page title is nil")
	}

	runtime.EventsEmit(a.wails, "PageDetails", PageDetails{
		ID:       *id,
		HasApi:   hasApi,
		LoggedIn: pi.LoggedIn(),
		Title:    *title,
		Type:     pi.Type(),
	})

	err := a.display(*name)
	if err != nil {
		return fmt.Errorf("error displaying page: %v", err)
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

func (a *App) startPageApi(pi PageInfo) error {
	name := pi.String()
	if name == nil {
		return fmt.Errorf("page name is nil")
	}
	url := pi.KeyUrl()
	if url == nil {
		return fmt.Errorf("page key url is nil")
	}

	if !a.producers.ApiP.Active(*name) {
		err := a.producers.ApiP.Start(*name, *url, 10*time.Second)
		if err != nil {
			return fmt.Errorf("error starting api: %v", err)
		}
		runtime.EventsEmit(a.wails, "ApiActive-"+*name, true)
	}

	return nil
}

// If page is inactivate, activate.
// If page is active, deactivate.
func (a *App) activatePage(pi PageInfo) error {
	name := pi.String()
	if name == nil {
		return fmt.Errorf("page name is nil")
	}
	url := pi.KeyUrl()
	if url == nil {
		return fmt.Errorf("page key url is nil")
	}

	a.pagesMu.Lock()
	page, exists := a.pages[*name]
	if !exists {
		page = &Page{
			active: false,
			apiSt:  &ApiState{},
			name:   *name,
		}
		a.pages[*name] = page
	}
	a.pagesMu.Unlock()

	page.activeMu.Lock()
	defer page.activeMu.Unlock()
	if page.active {
		if a.producers.ApiP.Active(*name) {
			err := a.producers.ApiP.Stop(*name)
			if err != nil {
				return fmt.Errorf("error stopping api: %v", err)
			}
		}

		page.displayingMu.Lock()
		if page.displaying {
			runtime.EventsEmit(a.wails, "PageActive", false)
		}
		page.displayingMu.Unlock()

		page.active = false
		return nil
	}
	page.active = true

	err := a.producers.ApiP.Start(*name, *url, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error starting api: %v", err)
	}
	runtime.EventsEmit(a.wails, "ApiActive-"+*name, true)

	err = a.display(*name)
	if err != nil {
		return fmt.Errorf("error displaying page: %v", err)
	}
	runtime.EventsEmit(a.wails, "PageActive", true)

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

	if a.producers.ApiP.Active(*name) {
		err := a.producers.ApiP.Stop(*name)
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

	runtime.EventsEmit(a.wails, "PageDetails", nil)

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error deleting account. Try again.")
	}
	runtime.EventsEmit(a.wails, "PageSideBarAccounts", list)

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

	if a.producers.ApiP.Active(*name) {
		err := a.producers.ApiP.Stop(*name)
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

	runtime.EventsEmit(a.wails, "PageDetails", nil)

	list, err := a.accountList()
	if err != nil {
		a.logError.Println("error getting account list:", err)
		return fmt.Errorf("Error deleting channel. Try again.")
	}
	runtime.EventsEmit(a.wails, "PageSideBarAccounts", list)

	return nil
}

func (a *App) display(name string) error {
	a.displayingMu.Lock()
	defer a.displayingMu.Unlock()

	if name == a.displaying {
		return nil
	}

	a.pagesMu.Lock()
	defer a.pagesMu.Unlock()

	if a.displaying != "" {
		displaying, exists := a.pages[a.displaying]
		a.displaying = ""
		if !exists {
			return fmt.Errorf("displaying page does not exist: %s", a.display)
		}
		displaying.displayingMu.Lock()
		displaying.displaying = false
		displaying.displayingMu.Unlock()
	}

	page, exists := a.pages[name]
	if !exists {
		page = &Page{
			active: false,
			apiSt:  &ApiState{},
			name:   name,
		}
		a.pages[name] = page
	}
	page.displayingMu.Lock()
	page.displaying = true
	page.displayingMu.Unlock()

	a.displaying = name

	err := a.updatePage(page)
	if err != nil {
		return fmt.Errorf("error updating page: %v", err)
	}

	return nil
}

func (a *App) updatePage(p *Page) error {
	if p == nil {
		return fmt.Errorf("page is nil")
	}

	//TODO check p.api == nil

	p.apiSt.respMu.Lock()
	defer p.apiSt.respMu.Unlock()

	p.displayingMu.Lock()
	if p.displaying {
		runtime.EventsEmit(a.wails, "PageActivity", p.apiSt.resp)
		p.activeMu.Lock()
		runtime.EventsEmit(a.wails, "PageActive", p.active)
		p.activeMu.Unlock()
	}
	p.displayingMu.Unlock()

	if p.apiSt.resp != nil {
		isLive := len(p.apiSt.resp.Livestreams) > 0
		runtime.EventsEmit(a.wails, "PageLive-"+p.name, isLive)
	}

	return nil
}

func (a *App) PageStatus(name string) {
	active := false
	isLive := false

	a.pagesMu.Lock()
	defer a.pagesMu.Unlock()
	page, exists := a.pages[name]
	if exists && page.apiSt != nil {
		page.apiSt.activeMu.Lock()
		active = page.apiSt.active
		page.apiSt.activeMu.Unlock()

		page.apiSt.respMu.Lock()
		if page.apiSt.resp != nil {
			isLive = len(page.apiSt.resp.Livestreams) > 0
		}
		page.apiSt.respMu.Unlock()
	}

	runtime.EventsEmit(a.wails, "ApiActive-"+name, active)
	runtime.EventsEmit(a.wails, "PageLive-"+name, isLive)
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

	if a.producers.ApiP.Active(*name) {
		err := a.producers.ApiP.Stop(*name)
		if err != nil {
			a.logError.Println("error stopping api:", err)
			return fmt.Errorf("Error updating account. Try again.")
		}
	}

	client := rumblelivestreamlib.Client{ApiKey: apiKey}
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

	if a.producers.ApiP.Active(*name) {
		err := a.producers.ApiP.Stop(*name)
		if err != nil {
			a.logError.Println("error stopping api:", err)
			return fmt.Errorf("Error updating channel. Try again.")
		}
	}

	client := rumblelivestreamlib.Client{ApiKey: apiKey}
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

func (a *App) DeleteChatbot(chatbot *models.Chatbot) error {
	if chatbot == nil || chatbot.ID == nil {
		return fmt.Errorf("Invalid chatbot. Try again.")
	}

	err := a.StopChatbotRules(chatbot.ID)
	if err != nil {
		a.logError.Println("error stopping chatbot rules before deleting chatbot")
		return fmt.Errorf("Error deleting chatbot. Could not stop running rules. Try Again.")
	}

	cb, err := a.services.ChatbotS.ByID(*chatbot.ID)
	if err != nil {
		a.logError.Println("error getting chatbot by ID:", err)
		return fmt.Errorf("Error deleting chatbot. Try again.")
	}
	if cb == nil {
		return fmt.Errorf("Chatbot does not exist.")
	}

	rules, err := a.services.ChatbotRuleS.ByChatbotID(*chatbot.ID)
	if err != nil {
		a.logError.Println("error getting chatbot rules by chatbot ID:", err)
		return fmt.Errorf("Error deleting chatbot. Try again.")
	}

	for _, rule := range rules {
		err = a.services.ChatbotRuleS.Delete(&rule)
		if err != nil {
			a.logError.Println("error deleting chatbot rule:", err)
			return fmt.Errorf("Error deleting chatbot. Try again.")
		}
	}

	err = a.services.ChatbotS.Delete(chatbot)
	if err != nil {
		a.logError.Println("error deleting chatbot:", err)
		return fmt.Errorf("Error deleting chatbot. Try again.")
	}

	list, err := a.chatbotList()
	if err != nil {
		a.logError.Println("error getting chatbot list:", err)
		return fmt.Errorf("Error deleting chatbot. Try again.")
	}
	runtime.EventsEmit(a.wails, "ChatbotList", list)

	return nil
}

func (a *App) NewChatbot(chatbot *models.Chatbot) error {
	if chatbot == nil || chatbot.Name == nil {
		return fmt.Errorf("Invalid chatbot. Try again.")
	}

	cb, err := a.services.ChatbotS.ByName(*chatbot.Name)
	if err != nil {
		a.logError.Println("error getting chatbot by name:", err)
		return fmt.Errorf("Error creating chatbot. Try again.")
	}
	if cb != nil {
		return fmt.Errorf("Chatbot name already exists.")
	}

	_, err = a.services.ChatbotS.Create(chatbot)
	if err != nil {
		a.logError.Println("error creating chatbot:", err)
		return fmt.Errorf("Error creating chatbot. Try again.")
	}

	list, err := a.chatbotList()
	if err != nil {
		a.logError.Println("error getting chatbot list:", err)
		return fmt.Errorf("Error creating chatbot. Try again.")
	}
	runtime.EventsEmit(a.wails, "ChatbotList", list)

	return nil
}

func (a *App) UpdateChatbot(chatbot *models.Chatbot) error {
	if chatbot == nil || chatbot.ID == nil || chatbot.Name == nil {
		return fmt.Errorf("Invalid chatbot. Try again.")
	}

	cb, err := a.services.ChatbotS.ByID(*chatbot.ID)
	if err != nil {
		a.logError.Println("error getting chatbot by ID:", err)
		return fmt.Errorf("Error updating chatbot. Try again.")
	}
	if cb == nil {
		return fmt.Errorf("Chatbot does not exist.")
	}

	cbByName, err := a.services.ChatbotS.ByName(*chatbot.Name)
	if err != nil {
		a.logError.Println("error getting chatbot by Name:", err)
		return fmt.Errorf("Error updating chatbot. Try again.")
	}
	if cbByName != nil && *cbByName.ID != *cb.ID {
		return fmt.Errorf("Chatbot name already exists.")
	}

	err = a.services.ChatbotS.Update(chatbot)
	if err != nil {
		a.logError.Println("error updating chatbot:", err)
		return fmt.Errorf("Error updating chatbot. Try again.")
	}

	list, err := a.chatbotList()
	if err != nil {
		a.logError.Println("error getting chatbot list:", err)
		return fmt.Errorf("Error updating chatbot. Try again.")
	}
	runtime.EventsEmit(a.wails, "ChatbotList", list)

	return nil
}

func (a *App) ChatbotList() ([]models.Chatbot, error) {
	list, err := a.chatbotList()
	if err != nil {
		a.logError.Println("error getting chatbot list:", err)
		return nil, fmt.Errorf("Error retrieving chatbots. Try restarting.")
	}

	return list, nil
}

func (a *App) chatbotList() ([]models.Chatbot, error) {
	list, err := a.services.ChatbotS.All()
	if err != nil {
		return nil, fmt.Errorf("error querying all chatbots: %v", err)
	}

	return list, err
}

func (a *App) ChatbotRules(chatbot *models.Chatbot) ([]chatbot.Rule, error) {
	if chatbot == nil || chatbot.ID == nil {
		return nil, fmt.Errorf("Invalid chatbot. Try again.")
	}

	rules, err := a.chatbotRules(*chatbot.ID)
	if err != nil {
		a.logError.Println("error getting chatbot rules:", err)
		return nil, fmt.Errorf("Error getting chatbot rules. Try again.")
	}

	return rules, nil
}

func (a *App) chatbotRules(chatbotID int64) ([]chatbot.Rule, error) {
	modelsRules, err := a.services.ChatbotRuleS.ByChatbotID(chatbotID)
	if err != nil {
		return nil, fmt.Errorf("error querying chatbot rules: %v", err)
	}

	rules := []chatbot.Rule{}
	for _, modelsRule := range modelsRules {
		rule := chatbot.Rule{
			ID:        modelsRule.ID,
			ChatbotID: modelsRule.ChatbotID,
		}

		if modelsRule.Parameters != nil {
			var params chatbot.RuleParameters
			err = json.Unmarshal([]byte(*modelsRule.Parameters), &params)
			if err != nil {
				return nil, fmt.Errorf("error un-marshaling chatbot rule parameters from json: %v", err)
			}

			rule.Parameters = &params
		}

		rule.Running = a.chatbot.Running(*rule.ChatbotID, *rule.ID)

		rule.Display = rule.Parameters.Message.FromText
		if rule.Parameters.Message.FromFile != nil {
			rule.Display = filepath.Base(rule.Parameters.Message.FromFile.Filepath)
		}

		rules = append(rules, rule)
	}

	chatbot.SortRules(rules)

	return rules, err
}

func (a *App) DeleteChatbotRule(rule *chatbot.Rule) error {
	if rule == nil || rule.ID == nil || rule.ChatbotID == nil {
		return fmt.Errorf("Invalid chatbot rule. Try again.")
	}

	mRule, err := rule.ToModelsChatbotRule()
	if err != nil {
		a.logError.Println("error converting chatbot.Rule into models.ChatbotRule:", err)
		return fmt.Errorf("Error deleting chatbot rule. Try again.")
	}

	err = a.chatbot.Stop(rule)
	if err != nil {
		a.logError.Println("error stopping chatbot rule:", err)
		return fmt.Errorf("Error deleting chatbot rule. Try again.")
	}

	err = a.services.ChatbotRuleS.Delete(mRule)
	if err != nil {
		a.logError.Println("error deleting chatbot rule:", err)
		return fmt.Errorf("Error deleting chatbot rule. Try again.")
	}

	rules, err := a.chatbotRules(*rule.ChatbotID)
	if err != nil {
		a.logError.Println("error getting chatbot rules:", err)
		return fmt.Errorf("Error deleting chatbot rule. Try again.")
	}
	runtime.EventsEmit(a.wails, "ChatbotRules", rules)

	return nil
}

func (a *App) NewChatbotRule(rule *chatbot.Rule) error {
	if rule == nil || rule.ChatbotID == nil || rule.Parameters == nil {
		return fmt.Errorf("Invalid chatbot rule. Try again.")
	}

	mRule, err := rule.ToModelsChatbotRule()
	if err != nil {
		a.logError.Println("error converting chatbot.Rule into models.ChatbotRule:", err)
		return fmt.Errorf("Error creating chatbot rule. Try again.")
	}

	_, err = a.services.ChatbotRuleS.Create(mRule)
	if err != nil {
		a.logError.Println("error creating chatbot rule:", err)
		return fmt.Errorf("Error creating chatbot rule. Try again.")
	}

	rules, err := a.chatbotRules(*rule.ChatbotID)
	if err != nil {
		a.logError.Println("error getting chatbot rules:", err)
		return fmt.Errorf("Error creating chatbot rule. Try again.")
	}
	runtime.EventsEmit(a.wails, "ChatbotRules", rules)

	return nil
}

func (a *App) RunChatbotRule(rule *chatbot.Rule) error {
	if rule == nil || rule.ChatbotID == nil {
		return fmt.Errorf("Invalid chatbot rule. Try again.")
	}

	mChatbot, err := a.services.ChatbotS.ByID(*rule.ChatbotID)
	if err != nil {
		a.logError.Println("error getting chatbot by ID:", err)
		return fmt.Errorf("Error running chatbot rule. Try again.")
	}
	if mChatbot == nil {
		return fmt.Errorf("Chatbot does not exist. Try again.")
	}
	if mChatbot.Url == nil {
		a.logError.Println("chatbot url is nil")
		return fmt.Errorf("Chatbot url is not set. Update url and try again.")
	}

	_, err = a.producers.ChatP.Start(*mChatbot.Url)
	if err != nil {
		a.logError.Println("error starting chat producer:", err)
		// TODO: send error to UI that chatbot URL could not be started
		//runtime.EventsEmit("Ch")
		return fmt.Errorf("Error connecting to chat. Try again.")
	}

	page := rule.Page()
	if page != nil {
		switch page.Prefix {
		case chatbot.PrefixAccount:
			acct, err := a.services.AccountS.ByUsername(page.Name)
			if err != nil {
				a.logError.Println("error getting account by username:", err)
				return fmt.Errorf("Error getting account to monitor. Check rule and try again.")
			}
			if acct == nil {
				return fmt.Errorf("Account to monitor does not exist. Check rule and try again.")
			}
			err = a.startPageApi(acct)
			if err != nil {
				a.logError.Println("error starting page api:", err)
				return fmt.Errorf("Error starting API for account in rule. Try again.")
			}
		case chatbot.PrefixChannel:
			channel, err := a.services.ChannelS.ByName(page.Name)
			if err != nil {
				a.logError.Println("error getting channel by name:", err)
				return fmt.Errorf("Error getting channel to monitor. Check rule and try again.")
			}
			if channel == nil {
				return fmt.Errorf("Channel to monitor does not exist. Check rule and try again.")
			}
			err = a.startPageApi(channel)
			if err != nil {
				a.logError.Println("error starting page api:", err)
				return fmt.Errorf("Error starting API for channel in rule. Try again.")
			}
		}
	}

	err = a.chatbot.Run(rule, *mChatbot.Url)
	if err != nil {
		a.logError.Println("error running chat bot rule:", err)
		return fmt.Errorf("Error running chatbot rule. Try again.")
	}

	return nil
}

func (a *App) StopChatbotRule(rule *chatbot.Rule) error {
	err := a.chatbot.Stop(rule)
	if err != nil {
		a.logError.Println("error stopping chat bot rule:", err)
		return fmt.Errorf("Error stopping chatbot rule. Try again.")
	}

	return nil
}

func (a *App) UpdateChatbotRule(rule *chatbot.Rule) error {
	if rule == nil || rule.ID == nil || rule.ChatbotID == nil {
		return fmt.Errorf("Invalid chatbot rule. Try again.")
	}

	mRule, err := rule.ToModelsChatbotRule()
	if err != nil {
		a.logError.Println("error converting chatbot.Rule into models.ChatbotRule:", err)
		return fmt.Errorf("Error updating chatbot rule. Try again.")
	}

	err = a.chatbot.Stop(rule)
	if err != nil {
		a.logError.Println("error stopping chatbot rule:", err)
		return fmt.Errorf("Error updating chatbot rule. Try again.")
	}

	err = a.services.ChatbotRuleS.Update(mRule)
	if err != nil {
		a.logError.Println("error updating chatbot rule:", err)
		return fmt.Errorf("Error updating chatbot rule. Try again.")
	}

	rules, err := a.chatbotRules(*rule.ChatbotID)
	if err != nil {
		a.logError.Println("error getting chatbot rules:", err)
		return fmt.Errorf("Error updating chatbot rule. Try again.")
	}
	runtime.EventsEmit(a.wails, "ChatbotRules", rules)

	return nil
}

func (a *App) RunChatbotRules(chatbotID *int64) error {
	if chatbotID == nil {
		return fmt.Errorf("Invalid chatbot. Try again.")
	}

	rules, err := a.chatbotRules(*chatbotID)
	if err != nil {
		a.logError.Println("error getting chatbot rules:", err)
		return fmt.Errorf("Error running chatbot rules. Try again.")
	}

	var errored bool
	for _, rule := range rules {
		if err = a.RunChatbotRule(&rule); err != nil {
			errored = true
		}
	}
	if errored {
		return fmt.Errorf("An error occurred while running rules. Check error log for details.")
	}

	return nil
}

func (a *App) StopChatbotRules(chatbotID *int64) error {
	if chatbotID == nil {
		return fmt.Errorf("Invalid chatbot. Try again.")
	}

	rules, err := a.chatbotRules(*chatbotID)
	if err != nil {
		a.logError.Println("error getting chatbot rules:", err)
		return fmt.Errorf("Error stopping chatbot rules. Try again.")
	}

	var errored bool
	for _, rule := range rules {
		if err = a.StopChatbotRule(&rule); err != nil {
			errored = true
		}
	}
	if errored {
		return fmt.Errorf("An error occurred while stopping rules. Check error log for details.")
	}

	return nil
}

func (a *App) OpenFileDialog() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		a.logError.Println("error getting home directory:", err)
		return "", fmt.Errorf("Error opening file explorer. Try again.")
	}

	filepath, err := runtime.OpenFileDialog(a.wails, runtime.OpenDialogOptions{DefaultDirectory: home})
	if err != nil {
		a.logError.Println("error opening file dialog:", err)
		return "", fmt.Errorf("Error opening file explorer. Try again.")
	}

	return filepath, err
}
