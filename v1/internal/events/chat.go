package events

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	rumblelivestreamlib "github.com/tylertravisty/rumble-livestream-lib-go"
)

type Chat struct {
	Livestream string
	Message    rumblelivestreamlib.ChatView
	Stop       bool
	Url        string
}

type chatProducer struct {
	cancel     context.CancelFunc
	cancelMu   sync.Mutex
	client     *rumblelivestreamlib.Client
	livestream string
	url        string
}

type chatProducerValFunc func(*chatProducer) error

func runChatProducerValFuncs(c *chatProducer, fns ...chatProducerValFunc) error {
	if c == nil {
		return fmt.Errorf("chat producer is nil")
	}

	for _, fn := range fns {
		err := fn(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func chatProducerRequireClient(c *chatProducer) error {
	if c.client == nil {
		return fmt.Errorf("client is nil")
	}

	return nil
}

type ChatProducer struct {
	Ch          chan Chat
	close       bool
	closeMu     sync.Mutex
	closeCh     chan bool
	logError    *log.Logger
	logInfo     *log.Logger
	producers   map[string]*chatProducer
	producersMu sync.Mutex
}

func NewChatProducer(logError *log.Logger, logInfo *log.Logger) *ChatProducer {
	return &ChatProducer{
		Ch:        make(chan Chat, 10),
		closeCh:   make(chan bool),
		logError:  logError,
		logInfo:   logInfo,
		producers: map[string]*chatProducer{},
	}
}

// func (cp *ChatProducer) Active(url string) bool {
// 	cp.producersMu.Lock()
// 	defer cp.producersMu.Unlock()
// 	_, active := cp.producers[url]

// 	return active
// }

func (cp *ChatProducer) Start(liveStreamUrl string) (string, error) {
	if liveStreamUrl == "" {
		return "", pkgErr("", fmt.Errorf("url is empty"))
	}

	cp.producersMu.Lock()
	defer cp.producersMu.Unlock()
	if producer, active := cp.producers[liveStreamUrl]; active {
		return producer.url, nil
	}

	client, err := rumblelivestreamlib.NewClient(rumblelivestreamlib.NewClientOptions{LiveStreamUrl: liveStreamUrl})
	if err != nil {
		return "", pkgErr("error creating new rumble client", err)
	}

	chatInfo, err := client.ChatInfo(false)
	if err != nil {
		return "", pkgErr("error getting chat info", err)
	}
	chatStreamUrl := chatInfo.StreamUrl()

	// cp.producersMu.Lock()
	// defer cp.producersMu.Unlock()
	// if _, active := cp.producers[chatStreamUrl]; active {
	// 	return chatStreamUrl, nil
	// }

	ctx, cancel := context.WithCancel(context.Background())
	producer := &chatProducer{
		cancel:     cancel,
		client:     client,
		livestream: liveStreamUrl,
		url:        chatStreamUrl,
	}
	// cp.producers[chatStreamUrl] = producer
	cp.producers[liveStreamUrl] = producer
	go cp.run(ctx, producer)

	return chatStreamUrl, nil
}

func (cp *ChatProducer) Stop(chatStreamUrl string) error {
	cp.producersMu.Lock()
	producer, exists := cp.producers[chatStreamUrl]
	if !exists {
		return pkgErr("", fmt.Errorf("producer does not exist for chat stream: %s", chatStreamUrl))
	}
	cp.producersMu.Unlock()

	producer.cancelMu.Lock()
	if producer.cancel != nil {
		producer.cancel()
	}
	producer.cancelMu.Unlock()

	return nil
}

func (cp *ChatProducer) run(ctx context.Context, producer *chatProducer) {
	err := runChatProducerValFuncs(
		producer,
		chatProducerRequireClient,
	)
	if err != nil {
		cp.logError.Println(pkgErr("invalid chat producer", err))
		return
	}

	// TODO: handle the case when restarting stream with possibly missing messages
	// Start new stream, make sure it's running, close old stream
	for {
		err = producer.client.StartChatStream(cp.handleChat(producer), cp.handleError(producer))
		if err != nil {
			cp.logError.Println(pkgErr("error starting chat stream", err))
			cp.stop(producer)
			return
		}

		timer := time.NewTimer(90 * time.Minute)
		select {
		case <-ctx.Done():
			timer.Stop()
			producer.client.StopChatStream()
			cp.stop(producer)
			return
		case <-timer.C:
			producer.client.StopChatStream()
		}
	}
}

func (cp *ChatProducer) handleChat(p *chatProducer) func(cv rumblelivestreamlib.ChatView) {
	return func(cv rumblelivestreamlib.ChatView) {
		if p == nil {
			return
		}

		cp.Ch <- Chat{Livestream: p.livestream, Message: cv, Url: p.url}
	}
}

func (cp *ChatProducer) handleError(p *chatProducer) func(err error) {
	return func(err error) {
		cp.logError.Println(pkgErr("chat stream returned error", err))
		p.cancelMu.Lock()
		if p.cancel != nil {
			p.cancel()
		}
		p.cancelMu.Unlock()
	}
}

func (cp *ChatProducer) stop(p *chatProducer) {
	if p == nil {
		return
	}

	cp.Ch <- Chat{Livestream: p.livestream, Stop: true, Url: p.url}

	cp.producersMu.Lock()
	delete(cp.producers, p.url)
	remaining := len(cp.producers)
	cp.producersMu.Unlock()

	cp.closeMu.Lock()
	if remaining == 0 && cp.close {
		select {
		case cp.closeCh <- true:
		default:
			break
		}
	}
	cp.closeMu.Unlock()

	return
}
