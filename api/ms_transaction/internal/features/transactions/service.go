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
	"strconv"
	"time"
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

	FindAll(
		ctx context.Context,
		page int,
		pageSize int,
	) ([]categories.CategoryDTO, error)
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
	Summary(
		ctx context.Context,
		query SummaryQuery,
	) (*SummaryDTO, error)
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
	userID := contexts.GetUser(ctx).GetID().String()

	var startKey, endKey, typeKey, categoryKey string
	if query.StartDate != nil {
		startKey = query.StartDate.Format(time.RFC3339)
	}
	if query.EndDate != nil {
		endKey = query.EndDate.Format(time.RFC3339)
	}
	if query.Type != nil {
		typeKey = strconv.Itoa(int(*query.Type))
	}
	if query.CategoryID != uuid.Nil() {
		categoryKey = query.CategoryID.String()
	}

	key := s.keyBuilder.BuildListKey(
		userID,
		startKey,
		endKey,
		typeKey,
		categoryKey,
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

	old, err := s.FindByID(ctx, model.ID)
	if err != nil {
		return err
	}

	if old.CategoryID == model.CategoryID {
		return s.we.Execute(ctx, func(ctx context.Context) error {
			return s.repo.Update(ctx, model)
		})
	}

	oldCategory, err := s.categoryClient.FindByID(ctx, old.CategoryID)
	if err != nil {
		return err
	}
	newCategory, err := s.categoryClient.FindByID(ctx, model.CategoryID)
	if err != nil {
		return err
	}

	err = s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Update(ctx, model)
	})

	if err != nil {
		return err
	}

	userAuth := contexts.GetUser(ctx)

	if oldCategory.GoalID != nil {
		if err := s.transactionProducer.PublishTransactionGoalDeleted(ctx,
			events.NewTransactionEvent(model.ID, 0, userAuth.GetID(), *oldCategory.GoalID)); err != nil {
			return err
		}
	}

	if newCategory.GoalID != nil {
		if err := s.transactionProducer.PublishTransactionGoalCreated(ctx,
			events.NewTransactionEvent(model.ID, model.Amount, userAuth.GetID(), *newCategory.GoalID)); err != nil {
			return err
		}
	}

	return nil
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

func (s *TransactionService) Summary(
	ctx context.Context,
	query SummaryQuery,
) (*SummaryDTO, error) {
	rows, err := s.repo.AggregateByCategory(ctx, query)
	if err != nil {
		return nil, err
	}

	categories, err := s.categoryClient.FindAll(ctx, 1, 100)
	if err != nil {
		return nil, err
	}

	type categoryInfo struct {
		Name string
		Type string
	}

	catMap := make(map[uuid.UUID]categoryInfo, len(categories))
	for _, c := range categories {
		catMap[c.ID] = categoryInfo{Name: c.Name, Type: c.Type}
	}

	var total, totalInput, totalOutput float64
	var count int64
	items := make([]SummaryItemDTO, len(rows))
	for i, row := range rows {
		total += row.Total
		count += row.Count

		info := catMap[row.CategoryID]

		switch info.Type {
		case "input":
			totalInput += row.Total
		case "output":
			totalOutput += row.Total
		}

		items[i] = SummaryItemDTO{
			CategoryID:   row.CategoryID,
			CategoryName: info.Name,
			Type:         info.Type,
			Total:        row.Total,
			Count:        row.Count,
		}
	}

	balance := totalInput - totalOutput

	return &SummaryDTO{
		Period: PeriodDTO{
			StartDate: query.StartDate,
			EndDate:   query.EndDate,
		},
		Total:       total,
		TotalInput:  totalInput,
		TotalOutput: totalOutput,
		Balance:     balance,
		Count:       count,
		ByCategory:  items,
	}, nil
}

func (s *TransactionService) DeleteByCategoryId(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.DeleteByCategoryId(ctx, id)
	})
}
