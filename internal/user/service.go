package user

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.FindByID(ctx, id)
}
