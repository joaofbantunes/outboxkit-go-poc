package polling

import (
	"context"

	"github.com/joaofbantunes/outboxkit-go-poc/core"
)

type BatchProduceResult struct {
	Ok []core.Message
}

type BatchProducer interface {
	Produce(ctx context.Context, key *core.Key, messages []core.Message) (BatchProduceResult, error)
}
