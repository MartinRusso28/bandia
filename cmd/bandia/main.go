package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MartinRusso28/bandia/internal/config"
	"github.com/MartinRusso28/bandia/internal/httpapi"
	"github.com/MartinRusso28/bandia/internal/store"
	"github.com/MartinRusso28/bandia/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, logger, os.Args[1:]); err != nil {
		logger.Error("bandia stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger, args []string) error {
	command := "serve"
	if len(args) > 0 {
		command = args[0]
	}
	if len(args) > 1 || (command != "serve" && command != "migrate") {
		return errors.New("usage: bandia [serve|migrate]")
	}
	cfg, err := config.Load(os.Getenv, command == "serve")
	if err != nil {
		return err
	}
	startup, cancel := context.WithTimeout(ctx, 15*time.Second)
	db, err := store.Open(startup, cfg.DatabaseURL)
	if err != nil {
		cancel()
		return err
	}
	defer db.Close()
	if command == "migrate" {
		defer cancel()
		if err := db.Migrate(startup); err != nil {
			return errors.New("migration failed; check database access and schema version")
		}
		logger.Info("database migrated")
		return nil
	}
	err = db.Ready(startup)
	cancel()
	if err != nil {
		return errors.New("schema unavailable or incompatible; run bandia migrate")
	}
	server := &http.Server{Addr: cfg.Address, Handler: httpapi.NewHandler(db, cfg.ManagerToken), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	workerCtx, stopWorker := context.WithCancel(ctx)
	workerDone := make(chan struct{})
	go func() { defer close(workerDone); worker.Run(workerCtx, db, logger) }()
	defer func() { stopWorker(); <-workerDone }()
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	logger.Info("starting HTTP server", "address", cfg.Address, "agents_available", false)
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.New("HTTP listener failed")
	case <-ctx.Done():
		logger.Info("shutting down")
		stopWorker()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return errors.New("HTTP shutdown timed out")
		}
		if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
			return errors.New("HTTP server stopped unexpectedly")
		}
		return nil
	}
}
