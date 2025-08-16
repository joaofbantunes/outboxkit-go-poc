package polling

import (
	"context"
	"fmt"
	"log"

	"github.com/joaofbantunes/outboxkit-go-poc/core"
)

type ProducePendingResult int

const (
	Ok ProducePendingResult = iota
	FetchError
	ProduceError
	PartialProduction
	CompleteError
)

type produceBatchResult int

const (
	moreAvailable produceBatchResult = iota
	allDone
	fetchError
	produceError
	partialProduction
	completeError
)

type Producer interface {
	ProducePending(ctx context.Context) ProducePendingResult
}

type DefaultProducer struct {
	key      core.Key
	fetcher  BatchFetcher
	producer BatchProducer
}

func NewProducer(key core.Key, fetcher BatchFetcher, producer BatchProducer) *DefaultProducer {
	return &DefaultProducer{
		key:      key,
		fetcher:  fetcher,
		producer: producer,
	}
}

func (p *DefaultProducer) ProducePending(ctx context.Context) ProducePendingResult {
	for {
		if ctx.Err() != nil {
			return Ok
		}

		result := p.produceBatch(ctx)
		switch result {
		case moreAvailable:
			continue
		case allDone:
			return Ok
		case fetchError:
			return FetchError
		case produceError:
			return ProduceError
		case partialProduction:
			return PartialProduction
		case completeError:
			return CompleteError
		default:
			panic(fmt.Sprintf("unknown produceBatchResult value: %d", result))
		}
	}
}

func (p *DefaultProducer) produceBatch(ctx context.Context) produceBatchResult {
	batchCtx, err := p.fetcher.FetchAndHold(ctx)
	if err != nil {
		log.Printf("error fetching batch: %v", err)
		return fetchError
	}

	defer batchCtx.Close(ctx)

	messages := batchCtx.Messages()
	if len(messages) == 0 {
		return allDone
	}

	result, err := p.producer.Produce(ctx, p.key, messages)
	if err != nil {
		log.Printf("error producing messages: %v", err)
		return produceError
	}

	err = batchCtx.Complete(ctx, result.Ok)
	if err != nil {
		// TODO: collect Ok messages for retry
		log.Printf("error completing batch: %v", err)
		return completeError
	}

	if len(result.Ok) < len(messages) {
		log.Printf("partial production: %d out of %d messages produced", len(result.Ok), len(messages))
		return partialProduction
	}

	hasNext, err := batchCtx.HasNext(ctx)
	if err != nil {
		log.Printf("error checking for next batch: %v", err)
		return fetchError
	}
	if hasNext {
		log.Printf("more messages available for key: %s", p.key)
		return moreAvailable
	}
	log.Printf("all messages processed for key: %s", p.key)
	return allDone
}
