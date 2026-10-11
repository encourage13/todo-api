package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/encourage13/todo-api/internal/auth"
	"github.com/encourage13/todo-api/internal/config"
	"github.com/encourage13/todo-api/internal/repository"
	"github.com/encourage13/todo-api/internal/service"
	"github.com/encourage13/todo-api/internal/transport/httpapi"
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

	logger.Info("success connection", "host", cfg.Postgres.Host)

	userRepo := repository.NewUserRepository(pool)
	todoRepo := repository.NewTodoRepository(pool)
	userSvc := service.NewUserService(userRepo)
	todoSvc := service.NewTodoService(todoRepo)
	tokenManager := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL)
	authHandler := httpapi.NewAuthHandler(userSvc, logger, tokenManager)
	todoHandler := httpapi.NewTodoHandler(todoSvc, logger)

	mux := http.NewServeMux()
	mux.Handle("/auth/", authHandler.Routes())
	requireAuth := httpapi.RequireAuth(tokenManager)
	protected := requireAuth(http.HandlerFunc(httpapi.Me))
	protectedTodo := requireAuth(todoHandler.Routes())

	mux.Handle("GET /me", protected)
	mux.Handle("/todos", protectedTodo)
	mux.Handle("/todos/", protectedTodo)
	srv := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           httpapi.JSONErrors(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("server starting", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}
