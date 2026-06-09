package todo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, in CreateInput) (*Todo, error)
	GetByID(ctx context.Context, userID, todoID uuid.UUID) (*Todo, error)
	List(ctx context.Context, p ListParams) ([]Todo, int, error)
	Update(ctx context.Context, userID, todoID uuid.UUID, in UpdateInput) (*Todo, error)
	Delete(ctx context.Context, userID, todoID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, in CreateInput) (*Todo, error) {
	now := time.Now()
	t := &Todo{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       in.Title,
		Description: in.Description,
		Completed:   false,
		DueDate:     in.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create todo: %w", err)
	}
	return t, nil
}

func (s *service) GetByID(ctx context.Context, userID, todoID uuid.UUID) (*Todo, error) {
	return s.repo.FindByID(ctx, userID, todoID)
}

func (s *service) List(ctx context.Context, p ListParams) ([]Todo, int, error) {
	return s.repo.List(ctx, p)
}

func (s *service) Update(ctx context.Context, userID, todoID uuid.UUID, in UpdateInput) (*Todo, error) {
	t, err := s.repo.FindByID(ctx, userID, todoID)
	if err != nil {
		return nil, err
	}

	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Description != nil {
		t.Description = in.Description
	}
	if in.Completed != nil {
		t.Completed = *in.Completed
	}
	if in.DueDate != nil {
		t.DueDate = in.DueDate
	}
	t.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}
	return t, nil
}

func (s *service) Delete(ctx context.Context, userID, todoID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, todoID)
}
