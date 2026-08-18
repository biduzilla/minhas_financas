package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"ms_auth/internal/core/contexts"
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/pkg/sqlformat"

	"github.com/google/uuid"
)

type RefreshTokenRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

type repository interface {
	FindByTokenHash(
		ctx context.Context,
		tokenHash string,
	) (*RefreshToken, error)

	RevokeAllByFamily(
		ctx context.Context,
		family uuid.UUID,
	) error

	Insert(
		ctx context.Context,
		model *RefreshToken,
	) error

	Update(ctx context.Context, model *RefreshToken) error

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) error
}

func NewRepository(
	db *sql.DB,
	logger *slog.Logger,
) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db:     db,
		logger: logger.WithGroup("db"),
	}
}

func (r *RefreshTokenRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*RefreshToken, error) {
	query := `
	select
		id,
		token_hash,
		user_id,
		expires_at,
		family,
		revoked,
		version,
		created_at,
		created_by,
		updated_at,
		updated_by
	from refresh_tokens
	where
		token_hash = $1
		and deleted = false
	`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	var model RefreshToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&model.ID,
		&model.TokenHash,
		&model.UserID,
		&model.ExpiresAt,
		&model.Family,
		&model.Revoked,
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

func (r *RefreshTokenRepository) RevokeAllByFamily(
	ctx context.Context,
	family uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
	update refresh_tokens 	
	set 
		revoked = true,
		updated_by = $2,
		updated_at = now()
	where 
		family = $1
		and deleted = false
	`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		panic("transaction necessary for this operation")
	}

	result, err := tx.ExecContext(ctx, query, family, userAuth.GetID())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return nil
	}

	return nil
}

func (r *RefreshTokenRepository) Insert(
	ctx context.Context,
	model *RefreshToken,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		insert into refresh_tokens (
			token_hash,
			user_id,
			expires_at,
			family,
			revoked,
			created_by
		) values (
			:token_hash,
			:user_id,
			:expires_at,
			:family,
			:revoked,
			:created_by
		) returning id, created_at, version
	`

	params := map[string]any{
		"token_hash": model.TokenHash,
		"user_id":    model.UserID,
		"expires_at": model.ExpiresAt,
		"family":     model.Family,
		"revoked":    model.Revoked,
		"created_by": userAuth.GetID(),
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
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) Update(ctx context.Context, model *RefreshToken) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		update refresh_tokens
		set 
			revoked = :revoked,
			updated_at = NOW(),
			updated_by = :user_id,
			version = version + 1
		where 
			id = :id
			and version = :version
			and deleted = false
		returning version
	`
	params := map[string]any{
		"revoked": model.Revoked,
		"user_id": userAuth.GetID(),
		"id":      model.ID,
		"version": model.Version,
	}

	query, args := sqlformat.NamedQuery(query, params)
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		panic("transaction necessary for this operation")
	}

	err := tx.QueryRowContext(ctx, query, args...).Scan(&model.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apiError.ErrEditConflict
		}
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		update refresh_tokens
		set 
			deleted = true,
			updated_at = NOW(),
			updated_by = :user_id,
			version = version + 1
		where 
			id = :id
			and deleted = false
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

	result, err := tx.ExecContext(ctx, query, args...)
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
