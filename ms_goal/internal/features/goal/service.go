package goal

import (
	"context"
	"ms_goal/internal/core/cache"
	"ms_goal/internal/core/contexts"
	"ms_goal/internal/core/domain/apiError"
	"ms_goal/internal/core/filters"
	"ms_goal/internal/core/messaging/events"
	"ms_goal/internal/core/validator"
	"uuid"
)

type GoalService struct {
	repo       repository
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
	we         WriteExecutor
	producers  goalProducer
	gtService  gtService
}

type goalProducer interface {
	PublishGoalCreated(
		ctx context.Context,
		event events.GoalEvent,
	) error

	PublishGoalDeleted(
		ctx context.Context,
		event events.GoalEvent,
	) error
}

type WriteExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type gtService interface {
	DeleteByGoalId(
		ctx context.Context,
		id uuid.UUID,
	) error
}
type service interface {
	FindByID(
		ctx context.Context,
		id uuid.UUID,
	) (*Goal, error)

	FindAll(
		ctx context.Context,
		search string,
		status GoalStatus,
		f filters.Filters,
	) ([]*Goal, filters.Metadata, error)

	Insert(
		ctx context.Context,
		model *Goal,
	) error

	Update(
		ctx context.Context,
		model *Goal,
	) error

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) error
}

func NewService(
	repo repository,
	cache cache.Cache,
	keyBuilder cache.KeyBuilder,
	we WriteExecutor,
	producers goalProducer,
	gtService gtService,
) *GoalService {
	return &GoalService{
		repo:       repo,
		cache:      cache,
		keyBuilder: keyBuilder,
		we:         we,
		producers:  producers,
		gtService:  gtService,
	}
}

func (s *GoalService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*Goal, error) {
	key := s.keyBuilder.BuildItemKey(id.String())

	return cache.FetchOrCache(ctx, s.cache, key, func() (*Goal, error) {
		return s.repo.FindByID(ctx, id)
	})
}

func (s *GoalService) FindAll(
	ctx context.Context,
	search string,
	status GoalStatus,
	f filters.Filters,
) ([]*Goal, filters.Metadata, error) {
	key := s.keyBuilder.BuildListKey(search, status, f.Page, f.PageSize, f.Sort)
	type listPayload struct {
		Models   []*Goal
		Metadata filters.Metadata
	}

	payload, err := cache.FetchOrCache(ctx, s.cache, key, func() (listPayload, error) {
		models, meta, err := s.repo.FindAll(ctx, search, status, f)
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

func (s *GoalService) Insert(
	ctx context.Context,
	model *Goal,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	err := s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, model)
	})
	if err != nil {
		return err
	}

	return s.producers.PublishGoalCreated(ctx, events.NewGoalEvent(
		model.ID, model.Name, contexts.GetUser(ctx).GetID(),
	))
}

func (s *GoalService) Update(
	ctx context.Context,
	model *Goal,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Update(ctx, model)
	})
}

func (s *GoalService) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := s.we.Execute(ctx, func(ctx context.Context) error {
		err := s.repo.DeleteById(ctx, id)
		if err != nil {
			return err
		}

		return s.gtService.DeleteByGoalId(ctx, id)
	})
	if err != nil {
		return err
	}

	return s.producers.PublishGoalDeleted(ctx, events.NewGoalEvent(
		id, "", contexts.GetUser(ctx).GetID(),
	))
}
