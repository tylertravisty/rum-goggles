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
	callers   map[string]*caller
	callersMu sync.Mutex
	ctx       context.Context
	display   string
	displayMu sync.Mutex
	logError  *log.Logger
	logInfo   *log.Logger
}

type caller struct {
	cancel     context.CancelFunc
	cancelMu   sync.Mutex
	display    bool
	displayMu  sync.Mutex
	interval   time.Duration
	name       string
	response   *rumblelivestreamlib.LivestreamResponse
	responseMu sync.Mutex
	url        string
}

type event struct {
	close bool
	err   error
	name  string
	resp  *rumblelivestreamlib.LivestreamResponse
}

func NewApi(logError *log.Logger, logInfo *log.Logger) *Api {
	return &Api{logError: logError, logInfo: logInfo}
}

func (a *Api) Response(name string) *rumblelivestreamlib.LivestreamResponse {
	a.callersMu.Lock()
	defer a.callersMu.Unlock()
	caller, exists := a.callers[name]
	if !exists {
		return nil
	}

	caller.responseMu.Lock()
	defer caller.responseMu.Unlock()

	copy := *caller.response
	return &copy
}

func (a *Api) Display(name string) error {
	a.displayMu.Lock()
	defer a.displayMu.Unlock()
	if name == a.display {
		return nil
	}

	a.callersMu.Lock()
	defer a.callersMu.Unlock()

	if a.display != "" {
		displaying, exists := a.callers[a.display]
		if !exists {
			return pkgErr("", fmt.Errorf("displaying caller does not exist: %s", a.display))
		}
		displaying.displayMu.Lock()
		displaying.display = false
		displaying.displayMu.Unlock()
		a.display = ""
	}

	caller, exists := a.callers[name]
	if !exists {
		// return pkgErr("", fmt.Errorf("caller does not exist: %s", name))
		runtime.EventsEmit(a.ctx, "PageActive", false)
		return nil
	}
	caller.displayMu.Lock()
	caller.display = true
	caller.displayMu.Unlock()

	a.display = name

	a.handleResponse(caller)

	return nil
}

func (a *Api) Startup(ctx context.Context) {
	a.ctx = ctx
	a.callers = map[string]*caller{}
}

func (a *Api) Shutdown() error {
	for _, caller := range a.callers {
		caller.cancelMu.Lock()
		if caller.cancel != nil {
			caller.cancel()
		}
		caller.cancelMu.Unlock()
	}

	return nil
}

func (a *Api) Start(name string, url string, interval time.Duration) error {
	if name == "" {
		return fmt.Errorf("name is empty")
	}
	if url == "" {
		return fmt.Errorf("url is empty")
	}

	a.callersMu.Lock()
	defer a.callersMu.Unlock()
	if _, active := a.callers[name]; active {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	caller := &caller{
		cancel:   cancel,
		interval: interval,
		name:     name,
		url:      url,
	}
	a.callers[name] = caller
	go a.run(ctx, caller)

	return nil
}

func (a *Api) run(ctx context.Context, caller *caller) {
	client := &rumblelivestreamlib.Client{StreamKey: caller.url}
	for {
		runtime.EventsEmit(a.ctx, "ApiActive-"+caller.name, true)
		caller.displayMu.Lock()
		if caller.display {
			runtime.EventsEmit(a.ctx, "PageActive", true)
		}
		caller.displayMu.Unlock()

		resp, err := a.query(client)
		if err != nil {
			a.logError.Println(pkgErr("error querying api", err))
			// runtime.EventsEmit(a.ctx, "ApiActive-"+caller.name, false)
			a.stop(caller)
			return
		}
		caller.responseMu.Lock()
		caller.response = resp
		caller.responseMu.Unlock()
		a.handleResponse(caller)

		timer := time.NewTimer(caller.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			// runtime.EventsEmit(a.ctx, "ApiActive-"+caller.name, false)
			a.stop(caller)
			return
		case <-timer.C:
		}
	}
}

func (a *Api) handleResponse(c *caller) {
	if c == nil {
		return
	}

	c.responseMu.Lock()
	defer c.responseMu.Unlock()
	if c.response == nil {
		return
	}

	c.displayMu.Lock()
	if c.display {
		runtime.EventsEmit(a.ctx, "PageActivity", c.response)
	}
	c.displayMu.Unlock()

	isLive := len(c.response.Livestreams) > 0
	runtime.EventsEmit(a.ctx, "PageLive-"+c.name, isLive)
}

func (a *Api) stop(c *caller) {
	if c == nil {
		return
	}

	runtime.EventsEmit(a.ctx, "ApiActive-"+c.name, false)
	c.displayMu.Lock()
	if c.display {
		c.display = false
		runtime.EventsEmit(a.ctx, "PageActive", false)
	}
	c.displayMu.Unlock()

	a.displayMu.Lock()
	if a.display == c.name {
		a.display = ""
	}
	a.displayMu.Unlock()

	a.callersMu.Lock()
	delete(a.callers, c.name)
	a.callersMu.Unlock()

	return
}

func (a *Api) Active(name string) bool {
	a.callersMu.Lock()
	defer a.callersMu.Unlock()
	_, active := a.callers[name]

	return active
}

func (a *Api) Stop(name string) error {
	a.callersMu.Lock()
	caller, exists := a.callers[name]
	if !exists {
		return pkgErr("", fmt.Errorf("caller does not exist: %s", name))
	}
	a.callersMu.Unlock()

	caller.cancelMu.Lock()
	if caller.cancel != nil {
		caller.cancel()
	}
	caller.cancelMu.Unlock()

	return nil
}

func (a *Api) query(client *rumblelivestreamlib.Client) (*rumblelivestreamlib.LivestreamResponse, error) {
	resp, err := client.Request()
	if err != nil {
		return nil, fmt.Errorf("error executing client request: %v", err)
	}

	return resp, nil
}
