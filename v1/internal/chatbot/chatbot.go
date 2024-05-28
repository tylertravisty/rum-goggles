package chatbot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tylertravisty/rum-goggles/v1/internal/events"
	"github.com/tylertravisty/rum-goggles/v1/internal/models"
	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type user struct {
	livestreams   map[string]*rumblelivestreamlib.Client
	livestreamsMu sync.Mutex
}

func (u *user) byLivestream(url string) *rumblelivestreamlib.Client {
	client, _ := u.livestreams[url]

	return client
}

type clients map[string]*user

func (c clients) byUsername(username string) *user {
	user, _ := c[username]
	return user
}

func (c clients) byUsernameLivestream(username string, url string) *rumblelivestreamlib.Client {
	user := c.byUsername(username)
	if user == nil {
		return nil
	}

	return user.byLivestream(url)
}

type followReceiver struct {
	apiCh  chan events.ApiFollower
	latest time.Time
}

type receiver struct {
	onCommand   map[string]map[int64]chan events.Chat
	onCommandMu sync.Mutex
	onFollow    map[int64]*followReceiver
	onFollowMu  sync.Mutex
	onRaid      map[int64]chan events.Chat
	onRaidMu    sync.Mutex
	onRant      map[int64]chan events.Chat
	onRantMu    sync.Mutex
	onSub       map[int64]chan events.Chat
	onSubMu     sync.Mutex
}

func newReceiver() *receiver {
	return &receiver{
		onCommand: map[string]map[int64]chan events.Chat{},
		onFollow:  map[int64]*followReceiver{},
		onRaid:    map[int64]chan events.Chat{},
		onRant:    map[int64]chan events.Chat{},
		onSub:     map[int64]chan events.Chat{},
	}
}

type Bot struct {
	runners   map[int64]*Runner
	runnersMu sync.Mutex
}

type Chatbot struct {
	accountS    models.AccountService
	bots        map[int64]*Bot
	botsMu      sync.Mutex
	chatbotS    models.ChatbotService
	clients     clients
	clientsMu   sync.Mutex
	logError    *log.Logger
	receivers   map[string]*receiver
	receiversMu sync.Mutex
	//runners     map[int64]*Runner
	// runnersMu sync.Mutex
	wails context.Context
}

func New(accountS models.AccountService, chatbotS models.ChatbotService, logError *log.Logger, wails context.Context) *Chatbot {
	return &Chatbot{
		accountS:  accountS,
		bots:      map[int64]*Bot{},
		chatbotS:  chatbotS,
		clients:   map[string]*user{},
		logError:  logError,
		receivers: map[string]*receiver{},
		// runners:   map[int64]*Runner{},
		wails: wails,
	}
}

// TODO: resetClient/updateClient
func (cb *Chatbot) addClient(username string, livestreamUrl string) (*rumblelivestreamlib.Client, error) {
	cb.clientsMu.Lock()
	defer cb.clientsMu.Unlock()

	u := cb.clients.byUsername(username)
	if u == nil {
		u = &user{
			livestreams: map[string]*rumblelivestreamlib.Client{},
		}
		cb.clients[username] = u
	}

	client := u.byLivestream(livestreamUrl)
	if client != nil {
		return client, nil
	}

	account, err := cb.accountS.ByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("error querying account by username: %v", err)
	}

	if account.Cookies == nil {
		return nil, fmt.Errorf("account cookies are nil")
	}

	var cookies []*http.Cookie
	err = json.Unmarshal([]byte(*account.Cookies), &cookies)
	if err != nil {
		return nil, fmt.Errorf("error un-marshaling cookie string: %v", err)
	}
	client, err = rumblelivestreamlib.NewClient(rumblelivestreamlib.NewClientOptions{Cookies: cookies, LiveStreamUrl: livestreamUrl})
	if err != nil {
		return nil, fmt.Errorf("error creating new client: %v", err)
	}

	_, err = client.ChatInfo(true)
	if err != nil {
		return nil, fmt.Errorf("error getting chat info for client: %v", err)
	}

	u.livestreamsMu.Lock()
	defer u.livestreamsMu.Unlock()
	u.livestreams[livestreamUrl] = client

	return client, nil
}

func (cb *Chatbot) Run(rule *Rule, url string) error {
	if rule == nil ||
		rule.ChatbotID == nil ||
		rule.ID == nil ||
		rule.Parameters == nil ||
		rule.Parameters.SendAs == nil {
		return pkgErr("", fmt.Errorf("invalid rule"))
	}

	stopped := cb.stopRunner(*rule.ChatbotID, *rule.ID)
	if stopped {
		// TODO: figure out better way to determine when running rule is cleaned up.
		// If rule was stopped, wait for everything to complete before running again.
		time.Sleep(1 * time.Second)
	}

	var err error
	client := cb.clients.byUsernameLivestream(rule.Parameters.SendAs.Username, url)
	if client == nil {
		client, err = cb.addClient(rule.Parameters.SendAs.Username, url)
		if err != nil {
			return pkgErr("error adding client", err)
		}
	}

	page := ""
	rulePage := rule.Page()
	if rulePage != nil {
		page = rulePage.Prefix + strings.ReplaceAll(rulePage.Name, " ", "")
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := &Runner{
		cancel: cancel,
		client: client,
		page:   page,
		rule:   *rule,
		wails:  cb.wails,
	}

	err = cb.initRunner(runner)
	if err != nil {
		return pkgErr("error initializing runner", err)
	}

	go cb.run(ctx, runner)

	return nil
}

func (cb *Chatbot) initRunner(runner *Runner) error {
	if runner == nil || runner.rule.ID == nil || runner.rule.ChatbotID == nil || runner.rule.Parameters == nil || runner.rule.Parameters.Trigger == nil {
		return fmt.Errorf("invalid runner")
	}

	channelID, err := runner.rule.Parameters.SendAs.ChannelIDInt()
	if err != nil {
		return fmt.Errorf("error converting channel ID to int: %v", err)
	}

	runner.channelIDMu.Lock()
	runner.channelID = channelID
	runner.channelIDMu.Unlock()

	switch {
	case runner.rule.Parameters.Trigger.OnCommand != nil:
		err = cb.initRunnerCommand(runner)
		if err != nil {
			return fmt.Errorf("error initializing command: %v", err)
		}
	case runner.rule.Parameters.Trigger.OnEvent != nil:
		err = cb.initRunnerEvent(runner)
		if err != nil {
			return fmt.Errorf("error initializing event: %v", err)
		}
	case runner.rule.Parameters.Trigger.OnTimer != nil:
		runner.run = runner.runOnTimer
	}

	// cb.runnersMu.Lock()
	// defer cb.runnersMu.Unlock()
	// cb.runners[*runner.rule.ID] = runner

	cb.botsMu.Lock()
	defer cb.botsMu.Unlock()
	bot, exists := cb.bots[*runner.rule.ChatbotID]
	if !exists {
		bot = &Bot{
			runners: map[int64]*Runner{},
		}

		cb.bots[*runner.rule.ChatbotID] = bot
	}

	bot.runnersMu.Lock()
	defer bot.runnersMu.Unlock()
	bot.runners[*runner.rule.ID] = runner

	return nil
}

func (cb *Chatbot) initRunnerCommand(runner *Runner) error {
	runner.run = runner.runOnCommand

	cmd := runner.rule.Parameters.Trigger.OnCommand.Command
	if cmd == "" || cmd[0] != '!' {
		return fmt.Errorf("invalid command")
	}

	chatCh := make(chan events.Chat, 10)
	runner.chatCh = chatCh

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()
	rcvr, exists := cb.receivers[runner.client.LiveStreamUrl]
	if !exists {
		rcvr = newReceiver()
		cb.receivers[runner.client.LiveStreamUrl] = rcvr
	}

	rcvr.onCommandMu.Lock()
	defer rcvr.onCommandMu.Unlock()
	chans, exists := rcvr.onCommand[cmd]
	if !exists {
		chans = map[int64]chan events.Chat{}
		rcvr.onCommand[cmd] = chans
	}
	chans[*runner.rule.ID] = chatCh

	return nil
}

func (cb *Chatbot) initRunnerEvent(runner *Runner) error {
	event := runner.rule.Parameters.Trigger.OnEvent
	switch {
	case event.FromAccount != nil:
		return cb.initRunnerEventFromAccount(runner)
	case event.FromChannel != nil:
		return cb.initRunnerEventFromChannel(runner)
	case event.FromLiveStream != nil:
		return cb.initRunnerEventFromLiveStream(runner)
	}

	return fmt.Errorf("runner event not supported")
}

func (cb *Chatbot) initRunnerEventFromAccount(runner *Runner) error {
	fromAccount := runner.rule.Parameters.Trigger.OnEvent.FromAccount
	switch {
	case fromAccount.OnFollow != nil:
		return cb.initRunnerEventFromAccountOnFollow(runner)
	}

	return fmt.Errorf("runner event not supported")
}

func (cb *Chatbot) initRunnerEventFromAccountOnFollow(runner *Runner) error {
	runner.run = runner.runOnEventFromAccountOnFollow

	apiCh := make(chan events.ApiFollower, 10)
	runner.apiCh = apiCh

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()
	rcvr, exists := cb.receivers[runner.page]
	if !exists {
		rcvr = newReceiver()
		cb.receivers[runner.page] = rcvr
	}

	// TODO: should I check if channel already exists, if so delete it?
	rcvr.onFollowMu.Lock()
	defer rcvr.onFollowMu.Unlock()
	rcvr.onFollow[*runner.rule.ID] = &followReceiver{apiCh, time.Now()}

	return nil
}

func (cb *Chatbot) initRunnerEventFromChannel(runner *Runner) error {
	fromChannel := runner.rule.Parameters.Trigger.OnEvent.FromChannel
	switch {
	case fromChannel.OnFollow != nil:
		return cb.initRunnerEventFromChannelOnFollow(runner)
	}

	return fmt.Errorf("runner event not supported")
}

func (cb *Chatbot) initRunnerEventFromChannelOnFollow(runner *Runner) error {
	runner.run = runner.runOnEventFromChannelOnFollow

	apiCh := make(chan events.ApiFollower, 10)
	runner.apiCh = apiCh

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()
	rcvr, exists := cb.receivers[runner.page]
	if !exists {
		rcvr = newReceiver()
		cb.receivers[runner.page] = rcvr
	}

	// TODO: should I check if channel already exists, if so delete it?
	rcvr.onFollowMu.Lock()
	defer rcvr.onFollowMu.Unlock()
	rcvr.onFollow[*runner.rule.ID] = &followReceiver{apiCh, time.Now()}

	return nil
}

func (cb *Chatbot) initRunnerEventFromLiveStream(runner *Runner) error {
	fromLiveStream := runner.rule.Parameters.Trigger.OnEvent.FromLiveStream
	switch {
	case fromLiveStream.OnRaid != nil:
		return cb.initRunnerEventFromLiveStreamOnRaid(runner)
	case fromLiveStream.OnRant != nil:
		return cb.initRunnerEventFromLiveStreamOnRant(runner)
	case fromLiveStream.OnSub != nil:
		return cb.initRunnerEventFromLiveStreamOnSub(runner)
	}

	return fmt.Errorf("runner event not supported")
}

func (cb *Chatbot) initRunnerEventFromLiveStreamOnRaid(runner *Runner) error {
	runner.run = runner.runOnEventFromLiveStreamOnRaid

	chatCh := make(chan events.Chat, 10)
	runner.chatCh = chatCh

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()
	rcvr, exists := cb.receivers[runner.client.LiveStreamUrl]
	if !exists {
		rcvr = newReceiver()
		cb.receivers[runner.client.LiveStreamUrl] = rcvr
	}

	// TODO: should I check if channel already exists, if so delete it?
	rcvr.onRaidMu.Lock()
	defer rcvr.onRaidMu.Unlock()
	rcvr.onRaid[*runner.rule.ID] = chatCh

	return nil
}

func (cb *Chatbot) initRunnerEventFromLiveStreamOnRant(runner *Runner) error {
	runner.run = runner.runOnEventFromLiveStreamOnRant

	chatCh := make(chan events.Chat, 10)
	runner.chatCh = chatCh

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()
	rcvr, exists := cb.receivers[runner.client.LiveStreamUrl]
	if !exists {
		rcvr = newReceiver()
		cb.receivers[runner.client.LiveStreamUrl] = rcvr
	}

	// TODO: should I check if channel already exists, if so delete it?
	rcvr.onRantMu.Lock()
	defer rcvr.onRantMu.Unlock()
	rcvr.onRant[*runner.rule.ID] = chatCh

	return nil
}

func (cb *Chatbot) initRunnerEventFromLiveStreamOnSub(runner *Runner) error {
	runner.run = runner.runOnEventFromLiveStreamOnSub

	chatCh := make(chan events.Chat, 10)
	runner.chatCh = chatCh

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()
	rcvr, exists := cb.receivers[runner.client.LiveStreamUrl]
	if !exists {
		rcvr = newReceiver()
		cb.receivers[runner.client.LiveStreamUrl] = rcvr
	}

	// TODO: should I check if channel already exists, if so delete it?
	rcvr.onSubMu.Lock()
	defer rcvr.onSubMu.Unlock()
	rcvr.onSub[*runner.rule.ID] = chatCh

	return nil
}

func (cb *Chatbot) run(ctx context.Context, runner *Runner) {
	if runner == nil || runner.rule.ID == nil || runner.run == nil {
		cb.logError.Println("invalid runner")
		return
	}

	runtime.EventsEmit(cb.wails, fmt.Sprintf("ChatbotRuleActive-%d", *runner.rule.ID), true)
	err := runner.run(ctx)
	if err != nil {
		prefix := fmt.Sprintf("chatbot runner for rule %d returned error:", *runner.rule.ID)
		cb.logError.Println(prefix, err)
		runtime.EventsEmit(cb.wails, fmt.Sprintf("ChatbotRuleError-%d", *runner.rule.ID), "Chatbot encountered an error while running this rule.")
	}

	err = cb.stop(&runner.rule)
	if err != nil {
		prefix := fmt.Sprintf("error stopping rule %d after runner returns:", *runner.rule.ID)
		cb.logError.Println(prefix, err)
		return
	}

	runtime.EventsEmit(cb.wails, fmt.Sprintf("ChatbotRuleActive-%d", *runner.rule.ID), false)
}

func (cb *Chatbot) Running(chatbotID int64, ruleID int64) bool {
	// cb.runnersMu.Lock()
	// defer cb.runnersMu.Unlock()
	// _, exists := cb.runners[id]
	// return exists

	cb.botsMu.Lock()
	defer cb.botsMu.Unlock()
	bot, exists := cb.bots[chatbotID]
	if !exists {
		return false
	}

	bot.runnersMu.Lock()
	defer bot.runnersMu.Unlock()
	_, exists = bot.runners[ruleID]
	return exists
}

func (cb *Chatbot) Stop(rule *Rule) error {
	err := cb.stop(rule)
	if err != nil {
		return pkgErr("", err)
	}

	return nil
}

func (cb *Chatbot) stop(rule *Rule) error {
	if rule == nil || rule.ID == nil || rule.ChatbotID == nil {
		return fmt.Errorf("invalid rule")
	}

	cb.stopRunner(*rule.ChatbotID, *rule.ID)

	return nil
}

func (cb *Chatbot) stopRunner(chatbotID int64, ruleID int64) bool {
	cb.botsMu.Lock()
	defer cb.botsMu.Unlock()
	bot, exists := cb.bots[chatbotID]
	if !exists {
		return false
	}

	bot.runnersMu.Lock()
	defer bot.runnersMu.Unlock()
	runner, exists := bot.runners[ruleID]
	if !exists {
		return false
	}

	stopped := true
	runner.stop()
	delete(bot.runners, ruleID)

	switch {
	case runner.rule.Parameters.Trigger.OnCommand != nil:
		err := cb.closeRunnerCommand(runner)
		if err != nil {
			cb.logError.Println("error closing runner command:", err)
		}
	case runner.rule.Parameters.Trigger.OnEvent != nil:
		err := cb.closeRunnerEvent(runner)
		if err != nil {
			cb.logError.Println("error closing runner event:", err)
		}
	}

	return stopped
}

func (cb *Chatbot) closeRunnerCommand(runner *Runner) error {
	if runner == nil || runner.rule.ID == nil || runner.rule.Parameters == nil || runner.rule.Parameters.Trigger == nil || runner.rule.Parameters.Trigger.OnCommand == nil {
		return fmt.Errorf("invalid runner command")
	}

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	rcvr, exists := cb.receivers[runner.client.LiveStreamUrl]
	if !exists {
		return fmt.Errorf("receiver for runner does not exist")
	}

	rcvr.onCommandMu.Lock()
	defer rcvr.onCommandMu.Unlock()

	cmd := runner.rule.Parameters.Trigger.OnCommand.Command
	chans, exists := rcvr.onCommand[cmd]
	if !exists {
		return fmt.Errorf("channel map for runner does not exist")
	}

	ch, exists := chans[*runner.rule.ID]
	if !exists {
		return fmt.Errorf("channel for runner does not exist")
	}

	close(ch)
	delete(chans, *runner.rule.ID)

	return nil
}

func (cb *Chatbot) closeRunnerEvent(runner *Runner) error {
	if runner == nil || runner.rule.ID == nil || runner.rule.Parameters == nil || runner.rule.Parameters.Trigger == nil || runner.rule.Parameters.Trigger.OnEvent == nil {
		return fmt.Errorf("invalid runner event")
	}

	switch {
	case runner.rule.Parameters.Trigger.OnEvent.FromAccount != nil:
		return cb.closeRunnerEventFromAccount(runner)
	case runner.rule.Parameters.Trigger.OnEvent.FromChannel != nil:
		return cb.closeRunnerEventFromChannel(runner)
	case runner.rule.Parameters.Trigger.OnEvent.FromLiveStream != nil:
		return cb.closeRunnerEventFromLiveStream(runner)
	}

	return fmt.Errorf("runner event not supported")
}

func (cb *Chatbot) closeRunnerEventFromAccount(runner *Runner) error {
	if runner == nil || runner.rule.ID == nil || runner.rule.Parameters == nil || runner.rule.Parameters.Trigger == nil || runner.rule.Parameters.Trigger.OnEvent == nil || runner.rule.Parameters.Trigger.OnEvent.FromAccount == nil {
		return fmt.Errorf("invalid runner event")
	}

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	rcvr, exists := cb.receivers[runner.page]
	if !exists {
		return fmt.Errorf("receiver for runner does not exist")
	}

	fromAccount := runner.rule.Parameters.Trigger.OnEvent.FromAccount
	switch {
	case fromAccount.OnFollow != nil:
		rcvr.onFollowMu.Lock()
		defer rcvr.onFollowMu.Unlock()
		followR, exists := rcvr.onFollow[*runner.rule.ID]
		if !exists {
			return fmt.Errorf("channel for runner does not exist")
		}
		close(followR.apiCh)
		delete(rcvr.onFollow, *runner.rule.ID)
	}

	return nil
}

func (cb *Chatbot) closeRunnerEventFromChannel(runner *Runner) error {
	if runner == nil || runner.rule.ID == nil || runner.rule.Parameters == nil || runner.rule.Parameters.Trigger == nil || runner.rule.Parameters.Trigger.OnEvent == nil || runner.rule.Parameters.Trigger.OnEvent.FromChannel == nil {
		return fmt.Errorf("invalid runner event")
	}

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	rcvr, exists := cb.receivers[runner.page]
	if !exists {
		return fmt.Errorf("receiver for runner does not exist")
	}

	fromChannel := runner.rule.Parameters.Trigger.OnEvent.FromChannel
	switch {
	case fromChannel.OnFollow != nil:
		rcvr.onFollowMu.Lock()
		defer rcvr.onFollowMu.Unlock()
		followR, exists := rcvr.onFollow[*runner.rule.ID]
		if !exists {
			return fmt.Errorf("channel for runner does not exist")
		}
		close(followR.apiCh)
		delete(rcvr.onFollow, *runner.rule.ID)
	}

	return nil
}

func (cb *Chatbot) closeRunnerEventFromLiveStream(runner *Runner) error {
	if runner == nil || runner.rule.ID == nil || runner.rule.Parameters == nil || runner.rule.Parameters.Trigger == nil || runner.rule.Parameters.Trigger.OnEvent == nil || runner.rule.Parameters.Trigger.OnEvent.FromLiveStream == nil {
		return fmt.Errorf("invalid runner event")
	}

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	rcvr, exists := cb.receivers[runner.client.LiveStreamUrl]
	if !exists {
		return fmt.Errorf("receiver for runner does not exist")
	}

	fromLiveStream := runner.rule.Parameters.Trigger.OnEvent.FromLiveStream
	switch {
	case fromLiveStream.OnRaid != nil:
		rcvr.onRaidMu.Lock()
		defer rcvr.onRaidMu.Unlock()
		ch, exists := rcvr.onRaid[*runner.rule.ID]
		if !exists {
			return fmt.Errorf("channel for runner does not exist")
		}
		close(ch)
		delete(rcvr.onRaid, *runner.rule.ID)
	case fromLiveStream.OnRant != nil:
		rcvr.onRantMu.Lock()
		defer rcvr.onRantMu.Unlock()
		ch, exists := rcvr.onRant[*runner.rule.ID]
		if !exists {
			return fmt.Errorf("channel for runner does not exist")
		}
		close(ch)
		delete(rcvr.onRant, *runner.rule.ID)
	case fromLiveStream.OnSub != nil:
		rcvr.onSubMu.Lock()
		defer rcvr.onSubMu.Unlock()
		ch, exists := rcvr.onSub[*runner.rule.ID]
		if !exists {
			return fmt.Errorf("channel for runner does not exist")
		}
		close(ch)
		delete(rcvr.onSub, *runner.rule.ID)
	}

	return nil
}

func (cb *Chatbot) HandleApi(event events.Api) {
	errs := cb.runApiFuncs(
		event,
		cb.handleApiFollow,
	)

	for _, err := range errs {
		cb.logError.Println("chatbot: error handling api event:", err)
	}
}

type apiFunc func(api events.Api) error

func (cb *Chatbot) runApiFuncs(api events.Api, fns ...apiFunc) []error {
	// TODO: validate api response?

	errs := []error{}
	for _, fn := range fns {
		err := fn(api)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (cb *Chatbot) handleApiFollow(api events.Api) error {
	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	rcvr, exists := cb.receivers[api.Name]
	if !exists {
		return nil
	}
	if rcvr == nil {
		return fmt.Errorf("receiver is nil for API: %s", api.Name)
	}

	rcvr.onFollowMu.Lock()
	defer rcvr.onFollowMu.Unlock()

	for _, runner := range rcvr.onFollow {
		latest := runner.latest
		for _, follower := range api.Resp.Followers.RecentFollowers {
			followedOn, err := time.Parse(time.RFC3339, follower.FollowedOn)
			// TODO: fix this in the API, not in the code
			followedOn = followedOn.Add(-4 * time.Hour)
			if err != nil {
				return fmt.Errorf("error parsing followed_on time: %v", err)
			}
			if followedOn.After(runner.latest) {
				if followedOn.After(latest) {
					latest = followedOn
				}
				runner.apiCh <- events.ApiFollower{Username: follower.Username}
			}
		}
		runner.latest = latest
	}

	return nil
}

func (cb *Chatbot) HandleChat(event events.Chat) {
	switch event.Message.Type {
	case rumblelivestreamlib.ChatTypeMessages:
		cb.handleMessage(event)
	}
}

func (cb *Chatbot) handleMessage(event events.Chat) {
	errs := cb.runMessageFuncs(
		event,
		cb.handleMessageCommand,
		cb.handleMessageEventRaid,
		cb.handleMessageEventRant,
		cb.handleMessageEventSub,
	)

	for _, err := range errs {
		cb.logError.Println("chatbot: error handling message:", err)
	}
}

func (cb *Chatbot) runMessageFuncs(chat events.Chat, fns ...messageFunc) []error {
	// TODO: validate message

	errs := []error{}
	for _, fn := range fns {
		err := fn(chat)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

type messageFunc func(chat events.Chat) error

func (cb *Chatbot) handleMessageCommand(chat events.Chat) error {
	if strings.Index(chat.Message.Text, "!") != 0 {
		return nil
	}

	words := strings.Split(chat.Message.Text, " ")
	cmd := words[0]

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	receiver, exists := cb.receivers[chat.Livestream]
	if !exists {
		return nil
	}
	if receiver == nil {
		return fmt.Errorf("receiver is nil for livestream: %s", chat.Livestream)
	}

	receiver.onCommandMu.Lock()
	defer receiver.onCommandMu.Unlock()
	runners, exist := receiver.onCommand[cmd]
	if !exist {
		return nil
	}

	for _, runner := range runners {
		runner <- chat
	}

	return nil
}

func (cb *Chatbot) handleMessageEventRaid(chat events.Chat) error {
	if !chat.Message.Raid {
		return nil
	}

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	receiver, exists := cb.receivers[chat.Livestream]
	if !exists {
		return nil
	}
	if receiver == nil {
		return fmt.Errorf("receiver is nil for livestream: %s", chat.Livestream)
	}

	receiver.onRaidMu.Lock()
	defer receiver.onRaidMu.Unlock()

	for _, runner := range receiver.onRaid {
		runner <- chat
	}

	return nil
}

func (cb *Chatbot) handleMessageEventRant(chat events.Chat) error {
	if chat.Message.Rant == 0 {
		return nil
	}

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	receiver, exists := cb.receivers[chat.Livestream]
	if !exists {
		return nil
	}
	if receiver == nil {
		return fmt.Errorf("receiver is nil for livestream: %s", chat.Livestream)
	}

	receiver.onRantMu.Lock()
	defer receiver.onRantMu.Unlock()

	for _, runner := range receiver.onRant {
		runner <- chat
	}

	return nil
}

func (cb *Chatbot) handleMessageEventSub(chat events.Chat) error {
	if !chat.Message.Sub {
		return nil
	}

	cb.receiversMu.Lock()
	defer cb.receiversMu.Unlock()

	receiver, exists := cb.receivers[chat.Livestream]
	if !exists {
		return nil
	}
	if receiver == nil {
		return fmt.Errorf("receiver is nil for livestream: %s", chat.Livestream)
	}

	receiver.onSubMu.Lock()
	defer receiver.onSubMu.Unlock()

	for _, runner := range receiver.onSub {
		runner <- chat
	}

	return nil
}
