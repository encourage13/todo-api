package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/encourage13/todo-api/internal/config"
	"github.com/encourage13/todo-api/internal/repository"
)

func main() {

	configPath := flag.String("config", "config/config.yaml", "path to config")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := run(logger, *configPath); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}

}

func run(logger *slog.Logger, configPath string) error {
	cfg, err := config.Load(configPath)

	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := repository.NewPostgresPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	logger.Info("success connection", "host", cfg.Postgres.Host, "addr", cfg.Server.Addr)

	return nil
}
