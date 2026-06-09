// Package main is the entry point for the todo-list API server.
//
//	@title			Todo List API
//	@version		1.0
//	@description	Production-grade todo list API with JWT authentication.
//
//	@contact.name	API Support
//	@contact.email	support@example.com
//
//	@license.name	MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter: Bearer <token>
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

	"github.com/user/todo-list/internal/app"
	"github.com/user/todo-list/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	srv, log, err := app.Build(ctx, cfg)
	if err != nil {
		slog.Error("build app", "error", err)
		os.Exit(1)
	}

	log.Info("server starting", "port", cfg.AppPort, "env", cfg.AppEnv)
	log.Info("swagger UI available", "url", "http://localhost:"+cfg.AppPort+"/swagger/index.html")

	go func() {
		if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	log.Info("server stopped")
}
