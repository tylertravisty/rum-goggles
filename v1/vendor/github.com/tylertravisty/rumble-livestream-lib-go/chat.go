package rumblelivestreamlib

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/r3labs/sse/v2"
	"github.com/tylertravisty/go-utils/random"
	"golang.org/x/net/html"
	"gopkg.in/cenkalti/backoff.v1"
)

const (
	ChatBadgeRecurringSubscription = "recurring_subscription"
	ChatBadgeLocalsSupporter       = "locals_supporter"

	ChatTypeInit           = "init"
	ChatTypeMessages       = "messages"
	ChatTypeMuteUsers      = "mute_users"
	ChatTypeDeleteMessages = "delete_messages"
	ChatTypeSubscriber     = "locals_supporter"
	ChatTypeRaiding        = "raid_confirmed"
	ChatTypePinMessage     = "pin_message"
	ChatTypeUnpinMessage   = "unpin_message"
)

type ChatInfo struct {
	ChannelID int
	ChatID    string
	Page      string
	UrlPrefix string
}

func (ci *ChatInfo) MessageUrl() string {
	return fmt.Sprintf("%s/chat/%s/message", ci.UrlPrefix, ci.ChatID)
}

func (ci *ChatInfo) StreamUrl() string {
	return fmt.Sprintf("%s/chat/%s/stream", ci.UrlPrefix, ci.ChatID)
}

func (c *Client) ChatInfo(reload bool) (*ChatInfo, error) {
	var ci *ChatInfo
	var err error

	if c.chatInfo == nil || reload {
		ci, err = c.getChatInfo()
		if err != nil {
			return nil, pkgErr("error getting chat info", err)
		}
	}

	c.chatInfo = ci
	return ci, nil
}

func (c *Client) getChatInfo() (*ChatInfo, error) {
	if c.LiveStreamUrl == "" {
		return nil, fmt.Errorf("stream url is empty")
	}

	resp, err := c.getWebpage(c.LiveStreamUrl)
	if err != nil {
		return nil, fmt.Errorf("error getting stream webpage: %v", err)
	}
	defer resp.Body.Close()

	var chatInfo *ChatInfo
	var page string
	r := bufio.NewReader(resp.Body)
	lineS, err := r.ReadString('\n')
	for err == nil {
		if strings.Contains(lineS, "RumbleChat(") {
			start := strings.Index(lineS, "RumbleChat(") + len("RumbleChat(")
			if start == -1 {
				return nil, fmt.Errorf("error finding chat function in webpage")
			}
			end := strings.Index(lineS[start:], ");")
			if end == -1 {
				return nil, fmt.Errorf("error finding end of chat function in webpage")
			}
			args := parseRumbleChatArgs(lineS[start : start+end])
			channelID, err := strconv.Atoi(args[5])
			if err != nil {
				return nil, fmt.Errorf("error converting channel ID argument string to int: %v", err)
			}
			chatInfo = &ChatInfo{ChannelID: channelID, ChatID: args[1], UrlPrefix: args[0]}
		} else if strings.Contains(lineS, "media-by--a") && strings.Contains(lineS, "author") {
			r := strings.NewReader(lineS)
			node, err := html.Parse(r)
			if err != nil {
				return nil, fmt.Errorf("error parsing html tag with page name: %v", err)
			}
			if node.FirstChild != nil && node.FirstChild.LastChild != nil && node.FirstChild.LastChild.FirstChild != nil {
				for _, attr := range node.FirstChild.LastChild.FirstChild.Attr {
					if attr.Key == "href" {
						page = attr.Val
					}
				}
			}
		}

		lineS, err = r.ReadString('\n')
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading line from stream webpage: %v", err)
	}
	if chatInfo == nil {
		return nil, fmt.Errorf("did not find RumbleChat function call")
	}

	chatInfo.Page = page
	return chatInfo, nil
}

func parseRumbleChatArgs(argsS string) []string {
	open := 0

	args := []string{}
	arg := []rune{}
	for _, c := range argsS {
		if c == ',' && open == 0 {
			args = append(args, trimRumbleChatArg(string(arg)))
			arg = []rune{}
		} else {
			if c == '[' {
				open = open + 1
			}
			if c == ']' {
				open = open - 1
			}

			arg = append(arg, c)
		}
	}
	if len(arg) > 0 {
		args = append(args, trimRumbleChatArg(string(arg)))
	}

	return args
}

func trimRumbleChatArg(arg string) string {
	return strings.Trim(strings.TrimSpace(arg), "\"")
}

type ChatMessage struct {
	Text string `json:"text"`
}

type ChatData struct {
	RequestID string      `json:"request_id"`
	Message   ChatMessage `json:"message"`
	Rant      *string     `json:"rant"`
	ChannelID *int        `json:"channel_id"`
}

type ChatRequest struct {
	Data ChatData `json:"data"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type ChatResponse struct {
	Errors []Error `json:"errors"`
}

func (c *Client) Chat(message string, channelID *int) error {
	if c.httpClient == nil {
		return pkgErr("", fmt.Errorf("http client is nil"))
	}

	if c.chatInfo == nil {
		ci, err := c.getChatInfo()
		if err != nil {
			return pkgErr("error getting chat info", err)
		}
		c.chatInfo = ci
	}

	requestID, err := random.String(32)
	if err != nil {
		return pkgErr("error generating request ID", err)
	}
	body := ChatRequest{
		Data: ChatData{
			RequestID: requestID,
			Message: ChatMessage{
				Text: message,
			},
			Rant:      nil,
			ChannelID: channelID,
		},
	}
	// if asChannel {
	// 	body.Data.ChannelID = &c.chatInfo.ChannelID
	// }

	bodyB, err := json.Marshal(body)
	if err != nil {
		return pkgErr("error marshaling request body into json", err)
	}

	resp, err := c.httpClient.Post(c.chatInfo.MessageUrl(), "application/json", bytes.NewReader(bodyB))
	if err != nil {
		return pkgErr("http Post request returned error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http Post response status not %s: %s", http.StatusText(http.StatusOK), resp.Status)
	}

	var cr ChatResponse
	err = json.NewDecoder(strings.NewReader(string(bodyB))).Decode(&cr)
	if err != nil {
		return fmt.Errorf("error decoding response body from server: %v", err)
	}

	if len(cr.Errors) != 0 {
		return fmt.Errorf("server returned an error: %s", cr.Errors[0].Message)
	}

	return nil
}

type ChatStream struct {
	sseClient *sse.Client
	sseEvent  chan *sse.Event
	stop      context.CancelFunc
}

type ChatEventChannel struct {
	ID       string `json:"id"`
	Image1   string `json:"image.1"`
	Link     string `json:"link"`
	Username string `json:"username"`
}

type ChatEventBlockData struct {
	Text string `json:"text"`
}

type ChatEventBlock struct {
	Data ChatEventBlockData `json:"data"`
	Type string             `json:"type"`
}

type ChatEventRant struct {
	Duration   int    `json:"duration"`
	ExpiresOn  string `json:"expires_on"`
	PriceCents int    `json:"price_cents"`
}

type ChatEventMessage struct {
	Blocks    []ChatEventBlock `json:"blocks"`
	ChannelID *int64           `json:"channel_id"`
	ID        string           `json:"id"`
	Rant      *ChatEventRant   `json:"rant"`
	Text      string           `json:"text"`
	Time      string           `json:"time"`
	UserID    string           `json:"user_id"`
}

type ChatEventUser struct {
	Badges     []string `json:"badges"`
	Color      string   `json:"color"`
	ID         string   `json:"id"`
	Image1     string   `json:"image.1"`
	IsFollower bool     `json:"is_follower"`
	Link       string   `json:"link"`
	Username   string   `json:"username"`
}

type ChatEventData struct {
	Channels []ChatEventChannel `json:"channels"`
	Messages []ChatEventMessage `json:"messages"`
	Users    []ChatEventUser    `json:"users"`
}

type ChatEvent struct {
	Data      ChatEventData `json:"data"`
	RequestID string        `json:"request_id"`
	Type      string        `json:"type"`
}

type ChatEventDataNoChannels struct {
	// TODO: change [][]string to [][]any and test
	Channels [][]string         `json:"channels"`
	Messages []ChatEventMessage `json:"messages"`
	Users    []ChatEventUser    `json:"users"`
}

type ChatEventNoChannels struct {
	Data      ChatEventDataNoChannels `json:"data"`
	RequestID string                  `json:"request_id"`
	Type      string                  `json:"type"`
}

func (c *Client) StartChatStream(handle func(cv ChatView), handleError func(err error)) error {
	c.chatStreamMu.Lock()
	defer c.chatStreamMu.Unlock()
	if c.chatStream != nil {
		return pkgErr("", fmt.Errorf("chat stream already started"))
	}
	sseEvent := make(chan *sse.Event)

	if c.chatInfo == nil {
		ci, err := c.getChatInfo()
		if err != nil {
			return pkgErr("error getting chat info", err)
		}
		c.chatInfo = ci
	}
	sseCl := sse.NewClient(c.chatInfo.StreamUrl())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	sseCl.ReconnectStrategy = backoff.WithContext(
		backoff.NewExponentialBackOff(),
		ctx,
	)

	err := sseCl.SubscribeChan("", sseEvent)
	if err != nil {
		cancel()
		return pkgErr(fmt.Sprintf("error subscribing to chat stream %s", c.chatInfo.StreamUrl()), err)
	}

	streamCtx, stop := context.WithCancel(context.Background())

	c.chatStream = &ChatStream{sseClient: sseCl, sseEvent: sseEvent, stop: stop}
	go startChatStream(streamCtx, sseEvent, handle, handleError)

	return nil
}

func (c *Client) StopChatStream() {
	c.chatStreamMu.Lock()
	defer c.chatStreamMu.Unlock()

	if c.chatStream == nil {
		return
	}
	// TODO: what order should these be in?
	if c.chatStream.sseClient != nil {
		c.chatStream.sseClient.Unsubscribe(c.chatStream.sseEvent)
	}
	if c.chatStream.stop != nil {
		c.chatStream.stop()
	}

	c.chatStream = nil
}

func startChatStream(ctx context.Context, event chan *sse.Event, handle func(cv ChatView), handleError func(err error)) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-event:
			if msg == nil {
				handleError(fmt.Errorf("received nil event"))
			} else {
				chats, err := parseEvent(msg.Data)
				if err != nil {
					handleError(err)
				} else {
					for _, chat := range chats {
						handle(chat)
					}
				}
			}
		}
	}
}

type ChatView struct {
	Badges      []string
	ChannelName string
	Color       string
	ImageUrl    string
	Init        bool
	IsFollower  bool
	Rant        int
	Text        string
	Time        time.Time
	Type        string
	Username    string
}

func parseEvent(event []byte) ([]ChatView, error) {
	var ce ChatEvent
	err := json.Unmarshal(event, &ce)
	if err != nil {
		var cenc ChatEventNoChannels
		errNC := json.Unmarshal(event, &cenc)
		if errNC != nil {
			return nil, fmt.Errorf("error un-marshaling event: %v", err)
		}

		ce.Data.Messages = cenc.Data.Messages
		ce.Data.Users = cenc.Data.Users
		ce.Type = cenc.Type
	}

	users := chatUsers(ce.Data.Users)
	channels := chatChannels(ce.Data.Channels)

	messages, err := parseMessages(ce.Type, ce.Data.Messages, users, channels)
	if err != nil {
		return nil, fmt.Errorf("error parsing messages: %v", err)
	}

	return messages, nil
}

func chatUsers(users []ChatEventUser) map[string]ChatEventUser {
	usersMap := map[string]ChatEventUser{}
	for _, user := range users {
		usersMap[user.ID] = user
	}

	return usersMap
}

func chatChannels(channels []ChatEventChannel) map[string]ChatEventChannel {
	channelsMap := map[string]ChatEventChannel{}
	for _, channel := range channels {
		channelsMap[channel.ID] = channel
	}

	return channelsMap
}

func parseMessages(eventType string, messages []ChatEventMessage, users map[string]ChatEventUser, channels map[string]ChatEventChannel) ([]ChatView, error) {
	views := []ChatView{}
	for _, message := range messages {
		var view ChatView
		user, exists := users[message.UserID]
		if !exists {
			return nil, fmt.Errorf("user ID does not exist: %s", message.UserID)
		}

		view.Badges = user.Badges
		view.Color = user.Color
		view.ImageUrl = user.Image1
		view.IsFollower = user.IsFollower
		if message.Rant != nil {
			view.Rant = message.Rant.PriceCents
		}
		view.Text = message.Text
		t, err := time.Parse(time.RFC3339, message.Time)
		if err != nil {
			return nil, fmt.Errorf("error parsing message time: %v", err)
		}
		view.Time = t
		view.Type = eventType
		view.Username = user.Username

		if message.ChannelID != nil {
			cid := strconv.Itoa(int(*message.ChannelID))
			channel, exists := channels[cid]
			if !exists {
				return nil, fmt.Errorf("channel ID does not exist: %d", *message.ChannelID)
			}

			view.ImageUrl = channel.Image1
			view.ChannelName = channel.Username
		}

		views = append(views, view)
	}

	return views, nil
}
