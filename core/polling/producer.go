package polling

import (
	"context"
	"fmt"
	"log/slog"

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
	logger   *slog.Logger
}

func NewProducer(key core.Key, fetcher BatchFetcher, producer BatchProducer, logProvider func(name string) *slog.Logger) *DefaultProducer {
	return &DefaultProducer{
		key:      key,
		fetcher:  fetcher,
		producer: producer,
		logger:   logProvider("producer"),
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
		p.logger.ErrorContext(ctx, "Error fetching batch", slog.Any("key", p.key), slog.Any("error", err))
		return fetchError
	}

	defer batchCtx.Close(ctx)

	messages := batchCtx.Messages()
	if len(messages) == 0 {
		return allDone
	}

	result, err := p.producer.Produce(ctx, p.key, messages)
	if err != nil {
		p.logger.ErrorContext(ctx, "Error producing messages", slog.Any("key", p.key), slog.Any("error", err))
		return produceError
	}

	err = batchCtx.Complete(ctx, result.Ok)
	if err != nil {
		// TODO: collect Ok messages for retry
		p.logger.ErrorContext(ctx, "Error completing batch", slog.Any("key", p.key), slog.Any("error", err))
		return completeError
	}

	if len(result.Ok) < len(messages) {
		p.logger.DebugContext(ctx, "Partial production occurred", slog.Any("key", p.key), slog.Int("produced", len(result.Ok)), slog.Int("total", len(messages)))
		return partialProduction
	}

	hasNext, err := batchCtx.HasNext(ctx)
	if err != nil {
		p.logger.ErrorContext(ctx, "Error checking for next batch", slog.Any("key", p.key), slog.Any("error", err))
		return fetchError
	}
	if hasNext {
		p.logger.DebugContext(ctx, "More messages available", slog.Any("key", p.key))
		return moreAvailable
	}
	p.logger.DebugContext(ctx, "All messages processed", slog.Any("key", p.key))
	return allDone
}
