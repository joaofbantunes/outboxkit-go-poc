package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
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

var logger = logProvider("main")

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

	logger.Info("Server starting", slog.String("addr", srv.Addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error", slog.String("addr", srv.Addr), slog.Any("error", err))
			os.Exit(1)
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
			logger.Error("Error with server shutdown", slog.Any("error", err))
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
			logger.ErrorContext(r.Context(), "Failed to close batch results", slog.Any("error", err))
		}
	}(batchResults)

	hasError := false
	for i := 0; i < count; i++ {
		_, err := batchResults.Exec()
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to insert message", slog.Int("index", i), slog.Any("error", err))
			hasError = true
		}
	}

	if hasError {
		http.Error(w, "Failed to insert some messages", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	logger.InfoContext(r.Context(), "Successfully inserted", slog.Int("count", count))
	trigger.OnNewMessages()
}

func migrateDB() {
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		logger.Error("Failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Error("postgres.WithInstance", slog.Any("error", err))
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"outboxkit_go_pg_sample",
		driver)

	if err != nil {
		logger.Error("migrate.NewWithDatabaseInstance", slog.Any("error", err))
		os.Exit(1)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Error("migrate.Up", slog.Any("error", err))
		os.Exit(1)
	}
}

func createPool() *pgxpool.Pool {
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		logger.Error("Failed to parse connection string", slog.Any("error", err))
		os.Exit(1)
	}

	pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("Failed to create connection pool", slog.Any("error", err))
		os.Exit(1)
	}

	return pool
}

func initOutbox(ctx context.Context) {
	key := pgPolling.NewDefaultKey()
	poller := corePolling.NewPoller(
		key,
		core.NewSystemTimeProvider(),
		corePolling.NewProducer(
			key,
			pgPolling.NewPgBatchFetcher(key, pool, logProvider),
			NewFakeBatchProducer(logProvider),
			logProvider,
		),
		logProvider,
	)
	trigger = poller.Trigger()
	poller.Start(ctx)
}

func logProvider(name string) *slog.Logger {
	// otelslog handler will receive a name, the built-in doesn't so adding the name as an attribute
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})).With(slog.String("source", name))
}
