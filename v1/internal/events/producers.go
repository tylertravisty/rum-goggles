package events

import (
	"fmt"
	"log"
)

type Producers struct {
	logError *log.Logger
	logInfo  *log.Logger
	ApiP     *ApiProducer
	ChatP    *ChatProducer
}

func (p *Producers) Startup() error {
	return nil
}

func (p *Producers) Shutdown() error {
	err := p.ApiP.Shutdown()
	if err != nil {
		return pkgErr("error shutting down api producer", err)
	}

	return nil
}

type ProducersInit func(*Producers) error

func NewProducers(inits ...ProducersInit) (*Producers, error) {
	var p Producers
	for _, init := range inits {
		err := init(&p)
		if err != nil {
			return nil, err
		}
	}

	return &p, nil
}

func WithLoggers(logError *log.Logger, logInfo *log.Logger) ProducersInit {
	return func(p *Producers) error {
		if logError == nil {
			return pkgErr("", fmt.Errorf("error logger is nil"))
		}
		p.logError = logError

		if logInfo == nil {
			return pkgErr("", fmt.Errorf("info logger is nil"))
		}
		p.logInfo = logInfo

		return nil
	}
}

func WithApiProducer() ProducersInit {
	return func(p *Producers) error {
		p.ApiP = NewApiProducer(p.logError, p.logInfo)

		return nil
	}
}

func WithChatProducer() ProducersInit {
	return func(p *Producers) error {
		p.ChatP = NewChatProducer(p.logError, p.logInfo)

		return nil
	}
}
