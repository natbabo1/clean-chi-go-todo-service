package app

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/user/todo-list/docs"
	"github.com/user/todo-list/internal/auth"
	"github.com/user/todo-list/internal/config"
	"github.com/user/todo-list/internal/platform/database"
	"github.com/user/todo-list/internal/platform/jwt"
	"github.com/user/todo-list/internal/platform/logger"
	"github.com/user/todo-list/internal/platform/password"
	"github.com/user/todo-list/internal/platform/validator"
	"github.com/user/todo-list/internal/todo"
	"github.com/user/todo-list/internal/user"
)

// Build wires all dependencies and returns a ready-to-start Server.
func Build(ctx context.Context, cfg *config.Config) (*Server, *slog.Logger, error) {
	log := logger.New(cfg.AppEnv)

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("connect database: %w", err)
	}

	hasher := password.NewBcryptHasher(12)
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessTokenTTL)
	v := validator.New()

	userRepo := user.NewPostgresRepository(pool)
	todoRepo := todo.NewPostgresRepository(pool)

	authSvc := auth.NewService(userRepo, hasher, jwtManager)
	userSvc := user.NewService(userRepo)
	todoSvc := todo.NewService(todoRepo)

	h := handlers{
		auth: auth.NewHandler(authSvc, v),
		user: user.NewHandler(userSvc),
		todo: todo.NewHandler(todoSvc, v),
	}

	router := buildRouter(log, jwtManager, cfg.CORSAllowedOrigins, h)
	srv := newServer(cfg.AppPort, router)

	return srv, log, nil
}
