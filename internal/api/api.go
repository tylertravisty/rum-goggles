package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Api struct {
	ctx             context.Context
	cancel          context.CancelFunc
	cancelMu        sync.Mutex
	querying        bool
	queryingMu      sync.Mutex
	queryInterval   time.Duration
	queryIntervalMu sync.Mutex
}

func NewApi() *Api {
	return &Api{queryInterval: 10 * time.Second}
}

func (a *Api) Startup(ctx context.Context) {
	a.ctx = ctx
	runtime.EventsOn(ctx, "StopQuery", func(optionalData ...interface{}) {
		a.Stop()
	})
}

func (a *Api) Start(url string) error {
	fmt.Println("Api.Start")
	if url == "" {
		return fmt.Errorf("empty stream key")
	}

	a.queryingMu.Lock()
	start := !a.querying
	a.querying = true
	a.queryingMu.Unlock()

	if start {
		fmt.Println("Starting querying")
		ctx, cancel := context.WithCancel(context.Background())
		a.cancelMu.Lock()
		a.cancel = cancel
		a.cancelMu.Unlock()
		a.start(ctx, url)
	} else {
		fmt.Println("Querying already started")
	}

	return nil
}

func (a *Api) Stop() {
	fmt.Println("stop querying")
	a.cancelMu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	a.cancelMu.Unlock()
}

func (a *Api) start(ctx context.Context, url string) {
	for {
		a.query(url)
		a.queryIntervalMu.Lock()
		interval := a.queryInterval
		a.queryIntervalMu.Unlock()
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			a.queryingMu.Lock()
			a.querying = false
			a.queryingMu.Unlock()
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (a *Api) query(url string) {
	fmt.Println("QueryAPI")
	client := rumblelivestreamlib.Client{StreamKey: url}
	resp, err := client.Request()
	if err != nil {
		// TODO: log error
		fmt.Println("client.Request err:", err)
		a.Stop()
		runtime.EventsEmit(a.ctx, "QueryResponseError", "Failed to query API")
		return
	}

	// resp := &rumblelivestreamlib.LivestreamResponse{}

	// resp.Followers.RecentFollowers = append(resp.Followers.RecentFollowers, rumblelivestreamlib.Follower{"tyler-follow", "2023-12-12T21:53:34-04:00"})
	// resp.Subscribers.RecentSubscribers = append(resp.Subscribers.RecentSubscribers, rumblelivestreamlib.Subscriber{"tyler-sub", "tyler-sub", 500, 5, "2023-12-14T21:53:34-04:00"})
	// resp.Subscribers.RecentSubscribers = append(resp.Subscribers.RecentSubscribers, rumblelivestreamlib.Subscriber{"tyler-sub", "tyler-sub", 500, 5, "2023-12-13T21:53:34-04:00"})
	// resp.Subscribers.RecentSubscribers = append(resp.Subscribers.RecentSubscribers, rumblelivestreamlib.Subscriber{"tyler-sub", "tyler-sub", 500, 5, "2023-11-13T21:53:34-04:00"})
	// resp.Livestreams = []rumblelivestreamlib.Livestream{
	// 	{
	// 		CreatedOn:   "2023-12-16T16:13:30+00:00",
	// 		WatchingNow: 4},
	// }
	runtime.EventsEmit(a.ctx, "QueryResponse", &resp)
}

// TODO: if start errors, send event
