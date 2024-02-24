package rumblelivestreamlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Follower struct {
	Username   string `json:"username"`
	FollowedOn string `json:"followed_on"`
}

type Followers struct {
	NumFollowers      int64      `json:"num_followers"`
	NumFollowersTotal int64      `json:"num_followers_total"`
	LatestFollower    Follower   `json:"latest_follower"`
	RecentFollowers   []Follower `json:"recent_followers"`
}

type Subscriber struct {
	User          string `json:"user"`
	Username      string `json:"username"`
	AmountCents   int64  `json:"amount_cents"`
	AmountDollars int64  `json:"amount_dollars"`
	SubscribedOn  string `json:"subscribed_on"`
}

type Subscribers struct {
	NumSubscribers    int64        `json:"num_subscribers"`
	LatestSubscriber  Subscriber   `json:"latest_subscriber"`
	RecentSubscribers []Subscriber `json:"recent_subscribers"`
}

type Category struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type Categories struct {
	Primary   Category `json:"primary"`
	Secondary Category `json:"secondary"`
}

type Badge string

type Message struct {
	Username  string  `json:"username"`
	Badges    []Badge `json:"badges"`
	Text      string  `json:"text"`
	CreatedOn string  `json:"created_on"`
}

type Rant struct {
	Message
	ExpiresOn     string `json:"expires_on"`
	AmountCents   int64  `json:"amount_cents"`
	AmountDollars int64  `json:"amount_dollars"`
}

type Chat struct {
	LatestMessage  Message   `json:"latest_message"`
	RecentMessages []Message `json:"recent_messages"`
	LatestRant     Rant      `json:"latest_rant"`
	RecentRants    []Rant    `json:"recent_rants"`
}

type Livestream struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	CreatedOn   string     `json:"created_on"`
	IsLive      bool       `json:"is_live"`
	Categories  Categories `json:"categories"`
	StreamKey   string     `json:"stream_key"`
	Likes       int64      `json:"likes"`
	Dislikes    int64      `json:"dislikes"`
	WatchingNow int64      `json:"watching_now"`
	Chat        Chat       `json:"chat"`
}

type LivestreamResponse struct {
	Now           int64        `json:"now"`
	Type          string       `json:"type"`
	UserID        string       `json:"user_id"`
	Username      string       `json:"username"`
	ChannelID     int64        `json:"channel_id"`
	ChannelName   string       `json:"channel_name"`
	MaxNumResults int64        `json:"max_num_results"`
	Followers     Followers    `json:"followers"`
	Subscribers   Subscribers  `json:"subscribers"`
	Livestreams   []Livestream `json:"livestreams"`
}

func (c *Client) Request() (*LivestreamResponse, error) {
	hcl := http.Client{Timeout: 30 * time.Second}
	resp, err := hcl.Get(c.StreamKey)
	if err != nil {
		return nil, pkgErr("http Get request returned error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, pkgErr(fmt.Sprintf("http response status not %s", http.StatusText(http.StatusOK)), fmt.Errorf("%s", resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgErr("error reading response body", err)
	}

	var lr LivestreamResponse
	err = json.Unmarshal(body, &lr)
	if err != nil {
		return nil, pkgErr("error unmarshaling response body", err)
	}

	return &lr, nil
}
