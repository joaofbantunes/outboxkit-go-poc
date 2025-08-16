package polling

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joaofbantunes/outboxkit-go-poc/core"
	"github.com/joaofbantunes/outboxkit-go-poc/core/polling"
)

type pgBatchFetcher struct {
	pool *pgxpool.Pool
}

func NewPgBatchFetcher(pool *pgxpool.Pool) polling.BatchFetcher {
	return &pgBatchFetcher{
		pool: pool,
	}
}

func (f *pgBatchFetcher) FetchAndHold(ctx context.Context) (polling.BatchContext, error) {

	query := `
			SELECT id, type, payload, created_at, trace_context
			FROM outbox_messages
			ORDER BY id
			LIMIT 10
			FOR UPDATE;
			`

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, query)

	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			return nil, errors.Join(err, rollbackErr)
		}
		return nil, err
	}

	defer rows.Close()

	var messages []core.Message
	for rows.Next() {
		var msg SampleMessage
		if err := rows.Scan(&msg.ID, &msg.Type, &msg.Payload, &msg.CreatedAt, &msg.TraceContext); err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				return nil, errors.Join(err, rollbackErr)
			}
			return nil, err
		}
		messages = append(messages, msg)
	}

	return &pgBatchContext{
		messages: messages,
		pool:     f.pool,
		tx:       tx,
	}, nil
}

type pgBatchContext struct {
	messages []core.Message
	pool     *pgxpool.Pool
	tx       pgx.Tx
}

func (c *pgBatchContext) Messages() []core.Message {
	return c.messages
}

func (c *pgBatchContext) Complete(ctx context.Context, ok []core.Message) error {
	query := `
			DELETE FROM outbox_messages
			WHERE id = ANY($1);
			`

	ids := make([]int64, len(ok))
	for i, msg := range ok {
		ids[i] = msg.(SampleMessage).ID
	}

	_, err := c.tx.Exec(ctx, query, ids)
	if err != nil {
		rollbackErr := c.tx.Rollback(ctx)
		if rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	return c.tx.Commit(ctx)
}

func (c *pgBatchContext) HasNext(ctx context.Context) (bool, error) {
	query := `
			SELECT EXISTS (
			SELECT 1
			FROM outbox_messages);
			`

	var exists bool
	err := c.pool.QueryRow(ctx, query).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *pgBatchContext) Close(ctx context.Context) {
	err := c.tx.Rollback(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		log.Printf("Error rolling back transaction: %v", err)
	}
}
