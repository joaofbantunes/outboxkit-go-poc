package polling

import (
	"context"
	"log/slog"
	"time"

	"github.com/joaofbantunes/outboxkit-go-poc/core"
)

type Poller struct {
	key      *core.Key
	tp       core.TimeProvider
	p        Producer
	listener *listener
	logger   *slog.Logger
}

func NewPoller(key *core.Key, tp core.TimeProvider, p Producer, logProvider func(name string) *slog.Logger) *Poller {
	return &Poller{
		key:      key,
		tp:       tp,
		p:        p,
		listener: newOutboxListener(),
		logger:   logProvider("poller").With(slog.String("key", key.String())),
	}
}

func (p *Poller) Trigger() OutboxTrigger {
	return p.listener
}

func (p *Poller) Start(ctx context.Context) {

	p.logger.DebugContext(ctx, "Starting poller")
	go p.loop(ctx)
}

func (p *Poller) loop(ctx context.Context) {
	duration := 30 * time.Second // TODO: make configurable
	ticker := p.tp.NewTicker(duration)
	defer ticker.Stop()

	// TODO: rethink the way ticker is being used, as with all the cases that are being handled,
	// it's being used incorrectly here (and possibly we need a timer instead)
	for {
		select {
		case <-p.listener.Chan():
			p.executeCycle(ctx)
		case <-ticker.Chan():
			p.executeCycle(ctx)
		case <-ctx.Done():
			p.logger.DebugContext(ctx, "Stopping poller")
			return
		}
	}
}

func (p *Poller) executeCycle(ctx context.Context) {
	result := p.p.ProducePending(ctx)
	switch result {
	case Ok:
		// TODO: Successfully produced pending items
		p.logger.DebugContext(ctx, "Successfully produced pending items")
	case FetchError:
		// TODO: Handle fetch error
		p.logger.DebugContext(ctx, "Fetch error occurred")
		panic(result)
	case ProduceError:
		// TODO: Handle produce error
		p.logger.DebugContext(ctx, "Produce error occurred")
		panic(result)
	case PartialProduction:
		// TODO: Handle partial production
		p.logger.DebugContext(ctx, "Partial production occurred")
		panic(result)
	case CompleteError:
		// TODO: Handle complete error
		p.logger.DebugContext(ctx, "Complete error occurred")
		panic(result)
	}
}
