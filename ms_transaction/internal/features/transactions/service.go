package transactions

import (
	"context"
	"ms_transaction/internal/core/httpclient/categories"
	"ms_transaction/internal/core/messaging/events"
	"shared/auth/contexts"
	"shared/auth/domain/apiError"
	"shared/cache"
	"shared/utils/filters"
	"shared/validator"
	"uuid"
)

type TransactionService struct {
	repo                repository
	cache               cache.Cache
	keyBuilder          cache.KeyBuilder
	we                  WriteExecutor
	categoryClient      categoryClient
	transactionProducer transactionProducer
}

type categoryClient interface {
	FindByID(
		ctx context.Context,
		id uuid.UUID,
	) (categories.CategoryDTO, error)
}

type WriteExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionProducer interface {
	PublishTransactionGoalCreated(
		ctx context.Context,
		event events.TransactionEvent,
	) error

	PublishTransactionGoalDeleted(
		ctx context.Context,
		event events.TransactionEvent,
	) error
}

type service interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	FindAll(ctx context.Context, query transactionQuery) ([]*Transaction, filters.Metadata, error)
	Insert(ctx context.Context, model *Transaction) error
	Update(ctx context.Context, model *Transaction) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	DeleteByCategoryId(ctx context.Context, id uuid.UUID) error
}

func NewService(
	repo repository,
	cache cache.Cache,
	keyBuilder cache.KeyBuilder,
	we WriteExecutor,
	categoryClient categoryClient,
	transactionProducer transactionProducer,
) *TransactionService {
	return &TransactionService{
		repo:                repo,
		cache:               cache,
		keyBuilder:          keyBuilder,
		we:                  we,
		categoryClient:      categoryClient,
		transactionProducer: transactionProducer,
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

	category, err := s.categoryClient.FindByID(ctx, model.CategoryID)
	if err != nil {
		return err
	}

	err = s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, model)
	})
	if err != nil {
		return err
	}

	if category.GoalID != nil {
		userAuth := contexts.GetUser(ctx)
		err := s.transactionProducer.PublishTransactionGoalCreated(ctx, events.NewTransactionEvent(
			model.ID, model.Amount, userAuth.GetID(), *category.GoalID,
		))
		if err != nil {
			return err
		}
	}

	return nil
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
	t, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	c, err := s.categoryClient.FindByID(ctx, t.CategoryID)
	if err != nil {
		return err
	}

	err = s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.DeleteById(ctx, id)
	})
	if err != nil {
		return err
	}

	if c.GoalID != nil {
		return s.transactionProducer.PublishTransactionGoalDeleted(
			ctx,
			events.NewTransactionEvent(
				id,
				0.0,
				t.UserID,
				*c.GoalID,
			))
	}

	return nil
}

func (s *TransactionService) DeleteByCategoryId(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.DeleteByCategoryId(ctx, id)
	})
}
