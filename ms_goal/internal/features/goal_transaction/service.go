package goaltransaction

import (
	"context"
	"ms_goal/internal/core/cache"
	"ms_goal/internal/core/domain/apiError"
	"ms_goal/internal/core/filters"
	"ms_goal/internal/core/validator"
	"uuid"
)

type GoalTransactionService struct {
	repo       repository
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
	we         WriteExecutor
}

type WriteExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type service interface {
	FindByID(
		ctx context.Context,
		id uuid.UUID,
	) (*GoalTransaction, error)

	FindAll(
		ctx context.Context,
		goalID uuid.UUID,
		f filters.Filters,
	) ([]*GoalTransaction, filters.Metadata, error)

	Insert(
		ctx context.Context,
		model *GoalTransaction,
	) error

	Update(
		ctx context.Context,
		model *GoalTransaction,
	) error

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) error

	DeleteByGoalId(
		ctx context.Context,
		id uuid.UUID,
	) error
}

func NewService(
	repo repository,
	cache cache.Cache,
	keyBuilder cache.KeyBuilder,
	we WriteExecutor,
) *GoalTransactionService {
	return &GoalTransactionService{
		repo:       repo,
		cache:      cache,
		keyBuilder: keyBuilder,
		we:         we,
	}
}

func (s *GoalTransactionService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*GoalTransaction, error) {
	key := s.keyBuilder.BuildItemKey(id.String())

	return cache.FetchOrCache(ctx, s.cache, key, func() (*GoalTransaction, error) {
		return s.repo.FindByID(ctx, id)
	})
}

func (s *GoalTransactionService) FindAll(
	ctx context.Context,
	goalId uuid.UUID,
	f filters.Filters,
) ([]*GoalTransaction, filters.Metadata, error) {
	key := s.keyBuilder.BuildListKey(goalId, f.Page, f.PageSize, f.Sort)
	type listPayload struct {
		Models   []*GoalTransaction
		Metadata filters.Metadata
	}

	payload, err := cache.FetchOrCache(ctx, s.cache, key, func() (listPayload, error) {
		models, meta, err := s.repo.FindAll(ctx, goalId, f)
		if err != nil {
			return listPayload{}, err
		}

		return listPayload{
			Models:   models,
			Metadata: meta,
		}, nil
	})

	if err != nil {
		return nil, filters.Metadata{}, err
	}

	return payload.Models, payload.Metadata, nil
}

func (s *GoalTransactionService) Insert(
	ctx context.Context,
	model *GoalTransaction,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, model)
	})
}

func (s *GoalTransactionService) Update(
	ctx context.Context,
	model *GoalTransaction,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Update(ctx, model)
	})
}

func (s *GoalTransactionService) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.DeleteById(ctx, id)
	})
}

func (s *GoalTransactionService) DeleteByGoalId(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.DeleteByGoalId(ctx, id)
	})
}
