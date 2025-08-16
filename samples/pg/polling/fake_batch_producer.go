package main

import (
	"context"
	"log"

	"github.com/joaofbantunes/outboxkit-go-poc/core"
	"github.com/joaofbantunes/outboxkit-go-poc/core/polling"
)

type fakeBatchProducer struct {
}

func NewFakeBatchProducer() polling.BatchProducer {
	return &fakeBatchProducer{}
}

func (p *fakeBatchProducer) Produce(ctx context.Context, key core.Key, messages []core.Message) (polling.BatchProduceResult, error) {
	// Simulate message production logic
	// In a real implementation, this would interact with a database or message queue
	for _, msg := range messages {
		// Here you would typically insert the message into the outbox table
		// For this fake implementation, we just log it
		log.Printf("Producing message %s for key %v", msg, key)
	}

	return polling.BatchProduceResult{Ok: messages}, nil
}
