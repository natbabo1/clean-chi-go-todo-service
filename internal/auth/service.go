package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/todo-list/internal/platform/jwt"
	"github.com/user/todo-list/internal/platform/password"
	"github.com/user/todo-list/internal/user"
)

type Service interface {
	Register(ctx context.Context, in RegisterInput) (*TokenResponse, error)
	Login(ctx context.Context, in LoginInput) (*TokenResponse, error)
}

type service struct {
	userRepo user.Repository
	hasher   password.Hasher
	jwt      *jwt.Manager
}

func NewService(userRepo user.Repository, hasher password.Hasher, jwtManager *jwt.Manager) Service {
	return &service{userRepo: userRepo, hasher: hasher, jwt: jwtManager}
}

func (s *service) Register(ctx context.Context, in RegisterInput) (*TokenResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	u := &user.User{
		ID:           uuid.New(),
		Email:        in.Email,
		PasswordHash: hash,
		Name:         in.Name,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.jwt.Generate(u.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &TokenResponse{AccessToken: token, User: user.ToResponse(u)}, nil
}

func (s *service) Login(ctx context.Context, in LoginInput) (*TokenResponse, error) {
	u, err := s.userRepo.FindByEmail(ctx, in.Email)
	if errors.Is(err, user.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := s.hasher.Compare(u.PasswordHash, in.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(u.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &TokenResponse{AccessToken: token, User: user.ToResponse(u)}, nil
}
