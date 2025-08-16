package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joaofbantunes/outboxkit-go-poc/core"
	corePolling "github.com/joaofbantunes/outboxkit-go-poc/core/polling"
	pgPolling "github.com/joaofbantunes/outboxkit-go-poc/pg/polling"
)

// TODO: move
var connStr = "postgres://user:pass@localhost:5432/outboxkit_go_pg_sample?sslmode=disable"

// TODO: move
var pool *pgxpool.Pool
var trigger corePolling.OutboxTrigger

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	migrateDB()
	pool = createPool()
	defer pool.Close()
	initOutbox(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /produce/{count}", produceHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Server starting on %s", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error with server: %v", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error with server shutdown: %v", err)
		}
	}()
	wg.Wait()
}

func produceHandler(w http.ResponseWriter, r *http.Request) {
	countStr := r.PathValue("count")

	count, err := strconv.Atoi(countStr)
	if err != nil {
		http.Error(w, "Invalid count parameter", http.StatusBadRequest)
		return
	}

	query := `
			INSERT INTO outbox_messages (type, payload, created_at, trace_context) 
			VALUES ($1, $2, $3, $4)
			`
	batch := &pgx.Batch{}

	for i := 0; i < count; i++ {

		batch.Queue(query,
			"sample",
			[]byte(gofakeit.HackeringVerb()),
			time.Now(),
			nil, // TODO
		)

	}
	batchResults := pool.SendBatch(
		r.Context(),
		batch,
	)

	defer func(batchResults pgx.BatchResults) {
		err := batchResults.Close()
		if err != nil {
			log.Printf("Failed to close batch results: %v", err)
		}
	}(batchResults)

	hasError := false
	for i := 0; i < count; i++ {
		_, err := batchResults.Exec()
		if err != nil {
			log.Printf("Failed to insert message %d: %v", i, err)
			hasError = true
		}
	}

	if hasError {
		http.Error(w, "Failed to insert some messages", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Inserted %d messages", count)
	trigger.OnNewMessages()
}

func migrateDB() {
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("postgres.WithInstance: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"outboxkit_go_pg_sample",
		driver)

	if err != nil {
		log.Fatalf("migrate.NewWithDatabaseInstance: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("m.Up: %v", err)
	}
}

func createPool() *pgxpool.Pool {
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Failed to parse pool config: %v", err)
	}

	pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}

	return pool
}

func initOutbox(ctx context.Context) {
	key := core.NewDefaultKey("pg_polling")
	poller := corePolling.NewPoller(
		key,
		core.NewSystemTimeProvider(),
		corePolling.NewProducer(
			key,
			pgPolling.NewPgBatchFetcher(pool),
			NewFakeBatchProducer(),
		),
	)
	trigger = poller.Trigger()
	poller.Start(ctx)
}
