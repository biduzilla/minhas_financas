package category

import (
	"context"
	"ms_category/internal/core/cache"
	"ms_category/internal/core/domain/apiError"
	"ms_category/internal/core/filters"
	"ms_category/internal/core/validator"
	"uuid"
)

type CategoryService struct {
	repo              repository
	cache             cache.Cache
	keyBuilder        cache.KeyBuilder
	we                WriteExecutor
	transactionClient transactionClient
}

type WriteExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionClient interface {
	DeleteByCategoryId(ctx context.Context, id uuid.UUID) error
}

func NewService(
	repo repository,
	cache cache.Cache,
	keyBuilder cache.KeyBuilder,
	we WriteExecutor,
	transactionClient transactionClient,
) *CategoryService {
	return &CategoryService{
		repo:              repo,
		cache:             cache,
		keyBuilder:        keyBuilder,
		we:                we,
		transactionClient: transactionClient,
	}
}

type service interface {
	FindByID(
		ctx context.Context,
		id uuid.UUID,
	) (*Category, error)

	FindAll(
		ctx context.Context,
		search string,
		f filters.Filters,
	) ([]*Category, filters.Metadata, error)

	Insert(
		ctx context.Context,
		model *Category,
	) error

	Update(
		ctx context.Context,
		model *Category,
	) error

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) error
}

func (s *CategoryService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*Category, error) {
	key := s.keyBuilder.BuildItemKey(id.String())

	return cache.FetchOrCache(ctx, s.cache, key, func() (*Category, error) {
		return s.repo.FindById(ctx, id)
	})
}

func (s *CategoryService) FindAll(
	ctx context.Context,
	search string,
	f filters.Filters,
) ([]*Category, filters.Metadata, error) {
	key := s.keyBuilder.BuildListKey(search, f.Page, f.PageSize, f.Sort)
	type listPayload struct {
		Models   []*Category
		Metadata filters.Metadata
	}

	payload, err := cache.FetchOrCache(ctx, s.cache, key, func() (listPayload, error) {
		models, meta, err := s.repo.FindAll(ctx, search, f)
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

func (s *CategoryService) Insert(
	ctx context.Context,
	model *Category,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, model)
	})
}

func (s *CategoryService) Update(
	ctx context.Context,
	model *Category,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Update(ctx, model)
	})
}

func (s *CategoryService) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		err := s.repo.DeleteById(ctx, id)
		if err != nil {
			return err
		}

		return s.transactionClient.DeleteByCategoryId(ctx, id)
	})
}
