package polling

import (
	"context"
	"log"
	"time"

	"github.com/joaofbantunes/outboxkit-go-poc/core"
)

type Poller struct {
	key core.Key
	tp  core.TimeProvider
	p   Producer
	l   *listener
}

func NewPoller(key core.Key, tp core.TimeProvider, p Producer) *Poller {
	return &Poller{
		key: key,
		tp:  tp,
		p:   p,
		l:   newOutboxListener(),
	}
}

func (p *Poller) Trigger() OutboxTrigger {
	return p.l
}

func (p *Poller) Start(ctx context.Context) {

	log.Printf("Starting poller with key: %s", p.key)
	go p.loop(ctx)
}

func (p *Poller) loop(ctx context.Context) {
	duration := 30 * time.Second // TODO: make configurable
	ticker := p.tp.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-p.l.Chan():
			p.executeCycle(ctx)
		case <-ticker.Chan():
			p.executeCycle(ctx)
		case <-ctx.Done():
			log.Printf("Ctx signaled, stopping poller with key: %s", p.key)
			return // Exit on context cancellation
		}
	}
}

func (p *Poller) executeCycle(ctx context.Context) {
	result := p.p.ProducePending(ctx)
	switch result {
	case Ok:
		// TODO: Successfully produced pending items
		log.Printf("Successfully produced pending items for key: %v", p.key)
	case FetchError:
		// TODO: Handle fetch error
		panic(result)
	case ProduceError:
		// TODO: Handle produce error
		panic(result)
	case PartialProduction:
		// TODO: Handle partial production
		panic(result)
	case CompleteError:
		// TODO: Handle complete error
		panic(result)
	}
}
