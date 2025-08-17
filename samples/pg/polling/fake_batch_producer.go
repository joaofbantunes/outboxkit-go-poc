package main

import (
	"context"
	"log/slog"

	"github.com/joaofbantunes/outboxkit-go-poc/core"
	"github.com/joaofbantunes/outboxkit-go-poc/core/polling"
)

type fakeBatchProducer struct {
	logger *slog.Logger
}

func NewFakeBatchProducer(logProvider func(name string) *slog.Logger) polling.BatchProducer {
	return &fakeBatchProducer{
		logProvider("fakeBatchProducer"),
	}
}

func (p *fakeBatchProducer) Produce(ctx context.Context, key *core.Key, messages []core.Message) (polling.BatchProduceResult, error) {
	for _, msg := range messages {
		p.logger.DebugContext(ctx, "Producing message", slog.Any("message", msg), slog.String("key", key.String()))
	}

	return polling.BatchProduceResult{Ok: messages}, nil
}
