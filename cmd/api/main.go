package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/encourage13/todo-api/internal/config"
)

func main() {

	configPath := flag.String("config", "config/config.yaml", "path to config")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("starting todo-api")

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("parse config", "error", err)
		os.Exit(1)
	}

	logger.Info("parsed config", "ip", cfg.Server.Addr, "host", cfg.Postgres.Host)

}
