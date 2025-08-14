package polling

import "context"

type ProducePendingResult int

const (
	Ok ProducePendingResult = iota
	FetchError
	ProduceError
	PartialProduction
	CompleteError
)

type Producer interface {
	ProducePending(ctx context.Context) ProducePendingResult
}

type ToNameProducer struct{}

func NewToNameProducer() *ToNameProducer {
	return &ToNameProducer{}
}

func (p *ToNameProducer) ProducePending(ctx context.Context) ProducePendingResult {
	// TODO: implement actual logic for producing pending messages
	return Ok
}
