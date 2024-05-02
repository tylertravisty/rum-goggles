package events

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
)

type Api struct {
	Name string
	Resp *rumblelivestreamlib.LivestreamResponse
	Stop bool
}

type apiProducer struct {
	cancel   context.CancelFunc
	cancelMu sync.Mutex
	interval time.Duration
	name     string
	url      string
}

type ApiProducer struct {
	Ch          chan Api
	close       bool
	closeMu     sync.Mutex
	closeCh     chan bool
	logError    *log.Logger
	logInfo     *log.Logger
	producers   map[string]*apiProducer
	producersMu sync.Mutex
}

func NewApiProducer(logError *log.Logger, logInfo *log.Logger) *ApiProducer {
	return &ApiProducer{
		Ch:        make(chan Api, 10),
		closeCh:   make(chan bool),
		logError:  logError,
		logInfo:   logInfo,
		producers: map[string]*apiProducer{},
	}
}

func (ap *ApiProducer) Active(name string) bool {
	ap.producersMu.Lock()
	defer ap.producersMu.Unlock()
	_, active := ap.producers[name]

	return active
}

func (ap *ApiProducer) Shutdown() error {
	wait := false
	ap.producersMu.Lock()
	if len(ap.producers) > 0 {
		ap.closeMu.Lock()
		ap.close = true
		ap.closeMu.Unlock()
	}
	for _, producer := range ap.producers {
		producer.cancelMu.Lock()
		if producer.cancel != nil {
			producer.cancel()
		}
		producer.cancelMu.Unlock()
	}
	ap.producersMu.Unlock()

	if wait {
		timer := time.NewTimer(3 * time.Second)
		select {
		case <-ap.closeCh:
			timer.Stop()
			close(ap.Ch)
		case <-timer.C:
			return pkgErr("", fmt.Errorf("not all producers were stopped"))
		}
	}

	return nil
}

func (ap *ApiProducer) Start(name string, url string, interval time.Duration) error {
	if name == "" {
		return pkgErr("", fmt.Errorf("name is empty"))
	}
	if url == "" {
		return pkgErr("", fmt.Errorf("url is empty"))
	}

	ap.producersMu.Lock()
	defer ap.producersMu.Unlock()
	if _, active := ap.producers[name]; active {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	producer := &apiProducer{
		cancel:   cancel,
		interval: interval,
		name:     name,
		url:      url,
	}
	ap.producers[name] = producer
	go ap.run(ctx, producer)

	return nil
}

func (ap *ApiProducer) Stop(name string) error {
	ap.producersMu.Lock()
	producer, exists := ap.producers[name]
	if !exists {
		return pkgErr("", fmt.Errorf("producer does not exist: %s", name))
	}
	ap.producersMu.Unlock()

	producer.cancelMu.Lock()
	if producer.cancel != nil {
		producer.cancel()
	}
	producer.cancelMu.Unlock()

	return nil
}

func (ap *ApiProducer) run(ctx context.Context, producer *apiProducer) {
	client := &rumblelivestreamlib.Client{ApiKey: producer.url}
	for {
		resp, err := apiQuery(client)
		if err != nil {
			ap.logError.Println(pkgErr("error querying api", err))
			ap.stop(producer)
			return
		}
		ap.handleResponse(producer, resp)

		timer := time.NewTimer(producer.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			ap.stop(producer)
			return
		case <-timer.C:
		}
	}
}

func (ap *ApiProducer) handleResponse(p *apiProducer, resp *rumblelivestreamlib.LivestreamResponse) {
	if p == nil || resp == nil {
		return
	}

	ap.Ch <- Api{Name: p.name, Resp: resp}
}

func apiQuery(client *rumblelivestreamlib.Client) (*rumblelivestreamlib.LivestreamResponse, error) {
	resp, err := client.Request()
	if err != nil {
		return nil, fmt.Errorf("error executing client request: %v", err)
	}

	return resp, nil
}

func (ap *ApiProducer) stop(p *apiProducer) {
	if p == nil {
		return
	}

	ap.Ch <- Api{Name: p.name, Stop: true}

	ap.producersMu.Lock()
	delete(ap.producers, p.name)
	remaining := len(ap.producers)
	ap.producersMu.Unlock()

	ap.closeMu.Lock()
	if remaining == 0 && ap.close {
		select {
		case ap.closeCh <- true:
		default:
			break
		}
	}
	ap.closeMu.Unlock()

	return
}
