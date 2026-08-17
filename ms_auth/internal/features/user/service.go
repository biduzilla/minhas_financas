package user

import (
	"context"
	"ms_auth/internal/core/cache"
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/core/filters"
	"ms_auth/internal/core/validator"

	"github.com/google/uuid"
)

type UserService struct {
	repo       repository
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
	we         WriteExecutor
}

type userService interface {
	SignUp(
		ctx context.Context,
		req CreateUserDTO,
	) (*User, error)
	FindAll(
		ctx context.Context,
		search string,
		f filters.Filters,
	) ([]*User, filters.Metadata, error)

	FindById(
		ctx context.Context,
		id uuid.UUID,
	) (*User, error)

	FindByEmail(
		ctx context.Context,
		email string,
	) (*User, error)

	Insert(
		ctx context.Context,
		model *User,
	) error

	Update(
		ctx context.Context,
		model *User,
	) error

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) error
}

type WriteExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewService(
	repo repository,
	cache cache.Cache,
	keyBuilder cache.KeyBuilder,
	we WriteExecutor,
) *UserService {
	return &UserService{
		repo:       repo,
		cache:      cache,
		keyBuilder: keyBuilder,
		we:         we,
	}
}

func (s *UserService) SignUp(
	ctx context.Context,
	req CreateUserDTO,
) (*User, error) {
	v := validator.New()
	validatePasswordPlaintext(v, string(req.Password))

	if !v.Valid() {
		return nil, apiError.NewValidationError(v.Errors)
	}
	user, err := req.ToModel()
	if err != nil {
		return nil, err
	}

	err = s.Insert(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, err
}

func (s *UserService) FindAll(
	ctx context.Context,
	search string,
	f filters.Filters,
) ([]*User, filters.Metadata, error) {
	key := s.keyBuilder.BuildListKey(search, f.Page, f.PageSize, f.Sort)
	type listPayload struct {
		Users    []*User
		Metadata filters.Metadata
	}
	payload, err := cache.FetchOrCache(ctx, s.cache, key, func() (listPayload, error) {
		users, meta, err := s.repo.FindAll(ctx, search, f)
		if err != nil {
			return listPayload{}, err
		}
		return listPayload{
			Users:    users,
			Metadata: meta,
		}, nil
	})
	if err != nil {
		return nil, filters.Metadata{}, err
	}

	return payload.Users, payload.Metadata, nil
}

func (s *UserService) FindById(
	ctx context.Context,
	id uuid.UUID,
) (*User, error) {
	key := s.keyBuilder.BuildItemKey(id.String())

	return cache.FetchOrCache(ctx, s.cache, key, func() (*User, error) {
		return s.repo.FindById(ctx, id)
	})
}

func (s *UserService) FindByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	key := s.keyBuilder.BuildItemKey(email)

	return cache.FetchOrCache(ctx, s.cache, key, func() (*User, error) {
		return s.repo.FindByEmail(ctx, email)
	})
}

func (s *UserService) Insert(
	ctx context.Context,
	model *User,
) error {
	v := validator.New()

	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, model)
	})
}

func (s *UserService) Update(
	ctx context.Context,
	model *User,
) error {
	v := validator.New()

	if model.Validate(v); !v.Valid() {
		return apiError.NewValidationError(v.Errors)
	}
	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Update(ctx, model)
	})
}

func (s *UserService) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.DeleteById(ctx, id)
	})
}
