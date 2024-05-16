package api

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Api struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cancelMu   sync.Mutex
	logError   *log.Logger
	logInfo    *log.Logger
	querying   bool
	queryingMu sync.Mutex
}

func NewApi(logError *log.Logger, logInfo *log.Logger) *Api {
	return &Api{logError: logError, logInfo: logInfo}
}

func (a *Api) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *Api) Start(url string, interval time.Duration) error {
	a.logInfo.Println("Api.Start")
	if url == "" {
		return fmt.Errorf("empty stream key")
	}

	a.queryingMu.Lock()
	start := !a.querying
	a.querying = true
	a.queryingMu.Unlock()

	if start {
		a.logInfo.Println("Start querying")
		ctx, cancel := context.WithCancel(context.Background())
		a.cancelMu.Lock()
		a.cancel = cancel
		a.cancelMu.Unlock()
		go a.start(ctx, url, interval)
	} else {
		a.logInfo.Println("Querying already started")
	}

	return nil
}

func (a *Api) Stop() {
	a.logInfo.Println("Stop querying")
	a.cancelMu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	a.cancelMu.Unlock()
}

func (a *Api) start(ctx context.Context, url string, interval time.Duration) {
	for {
		a.query(url)
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
	// a.logInfo.Println("QueryAPI")
	client := rumblelivestreamlib.Client{ApiKey: url}
	resp, err := client.Request()
	if err != nil {
		a.logError.Println("api: error executing client request:", err)
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
