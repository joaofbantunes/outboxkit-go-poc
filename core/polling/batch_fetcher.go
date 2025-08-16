package polling

import (
	"context"

	"github.com/joaofbantunes/outboxkit-go-poc/core"
)

type BatchFetcher interface {
	FetchAndHold(ctx context.Context) (BatchContext, error)
}

type BatchContext interface {
	Messages() []core.Message
	Complete(ctx context.Context, ok []core.Message) error
	HasNext(ctx context.Context) (bool, error)
	Close(ctx context.Context)
}
