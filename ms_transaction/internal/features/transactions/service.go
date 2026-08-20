package transactions

import (
	"context"
	"ms_transaction/internal/core/cache"
	"ms_transaction/internal/core/domain/apiError"
	"ms_transaction/internal/core/filters"
	"ms_transaction/internal/core/validator"
	"uuid"
)

type TransactionService struct {
	repo       repository
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
	we         WriteExecutor
}

type WriteExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type service interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	FindAll(ctx context.Context, query transactionQuery) ([]*Transaction, filters.Metadata, error)
	Insert(ctx context.Context, model *Transaction) error
	Update(ctx context.Context, model *Transaction) error
	DeleteById(ctx context.Context, id uuid.UUID) error
}

func NewService(
	repo repository,
	cache cache.Cache,
	keyBuilder cache.KeyBuilder,
	we WriteExecutor,
) *TransactionService {
	return &TransactionService{
		repo:       repo,
		cache:      cache,
		keyBuilder: keyBuilder,
		we:         we,
	}
}

func (s *TransactionService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*Transaction, error) {
	key := s.keyBuilder.BuildItemKey(id.String())

	return cache.FetchOrCache(ctx, s.cache, key, func() (*Transaction, error) {
		return s.repo.FindById(ctx, id)
	})
}

func (s *TransactionService) FindAll(
	ctx context.Context,
	query transactionQuery,
) ([]*Transaction, filters.Metadata, error) {
	key := s.keyBuilder.BuildListKey(
		query.StartDate,
		query.EndDate,
		query.Type,
		query.CategoryID,
		query.Filters.Page,
		query.Filters.PageSize,
		query.Filters.Sort,
	)

	type listPayload struct {
		Models   []*Transaction
		Metadata filters.Metadata
	}

	payload, err := cache.FetchOrCache(ctx, s.cache, key, func() (listPayload, error) {
		models, meta, err := s.repo.FindAll(ctx, query)
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

func (s *TransactionService) Insert(
	ctx context.Context,
	model *Transaction,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, model)
	})
}

func (s *TransactionService) Update(
	ctx context.Context,
	model *Transaction,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Update(ctx, model)
	})
}

func (s *TransactionService) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.DeleteById(ctx, id)
	})
}
