package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"ms_auth/internal/core/contexts"
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/pkg/sqlformat"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(
	db *sql.DB,
	logger *slog.Logger,
) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.WithGroup("db"),
	}
}

func parseConstraintError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Constraint == "uniq_email" {
		return apiError.ValidationAlreadyExists("email")
	}
	return err
}

func (r *UserRepository) FindById(
	ctx context.Context,
	id uuid.UUID,
) (*User, error) {
	query := `
	select 
		id,
		email,
		name,
		password_hash,
		activated,
		version,
		created_at,
		created_by,
		updated_at,
		updated_by
	from users 
	where 
		deleted = false
		and id = $1
	`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	var model User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID,
		&model.Email,
		&model.PasswordHash,
		&model.Activated,
		&model.Version,
		&model.CreatedAt,
		&model.CreatedBy,
		&model.UpdatedAt,
		&model.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiError.ErrRecordNotFound
		}
		return nil, err
	}

	return &model, nil
}

func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	query := `
	select 
		id,
		email,
		name,
		password_hash,
		activated,
		version,
		created_at,
		created_by,
		updated_at,
		updated_by
	from users 
	where 
		deleted = false
		and email = $1
	`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	var model User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&model.ID,
		&model.Email,
		&model.PasswordHash,
		&model.Activated,
		&model.Version,
		&model.CreatedAt,
		&model.CreatedBy,
		&model.UpdatedAt,
		&model.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiError.ErrRecordNotFound
		}
		return nil, err
	}

	return &model, nil
}

func (r *UserRepository) Insert(
	ctx context.Context,
	model *User,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		insert into users (
			email,
			name,
			password_hash,
			created_at,
		) values (
			:email,
			:name,
			:password_hash,
			:created_by,
		) returning id, created_at, version
	`

	params := map[string]any{
		"email":         model.Email,
		"name":          model.Name,
		"password_hash": model.PasswordHash,
		"created_by":    userAuth.GetID(),
	}

	query, args := sqlformat.NamedQuery(query, params)
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		panic("transaction necessary for this operation")
	}

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&model.ID,
		&model.CreatedAt,
		&model.Version,
	)

	if err != nil {
		return parseConstraintError(err)
	}

	return nil
}

func (r *UserRepository) Update(
	ctx context.Context,
	model *User,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		update users
		set
			email = :email,
			name = :name,
			updated_at = NOW(),
			updated_by = :user_id,
			version = version + 1
		WHERE 
			id = :id 
			AND version = :version 
			AND deleted = false
        RETURNING version
	`

	params := map[string]any{
		"name":    model.Name,
		"email":   model.Email,
		"id":      model.ID,
		"version": model.Version,
		"user_id": userAuth.GetID(),
	}

	query, args := sqlformat.NamedQuery(query, params)
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		panic("transaction necessary for this operation")
	}

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&model.ID,
		&model.CreatedAt,
		&model.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apiError.ErrEditConflict
		}
		return parseConstraintError(err)
	}

	return nil
}

func (r *UserRepository) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
        UPDATE users
        SET deleted = true, updated_at = NOW(), updated_by = :user_id, version = version + 1
        WHERE id = :id AND deleted = false
        RETURNING id
    `
	params := map[string]any{
		"id":      id,
		"user_id": userAuth.GetID(),
	}

	query, args := sqlformat.NamedQuery(query, params)
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		panic("transaction necessary for this operation")
	}

	result, err := tx.ExecContext(ctx, query, args)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apiError.ErrRecordNotFound
	}

	return nil
}
