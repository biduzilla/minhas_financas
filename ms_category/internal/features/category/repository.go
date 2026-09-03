package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"ms_category/internal/core/contexts"
	"ms_category/internal/core/domain/apiError"
	"ms_category/internal/core/filters"
	"ms_category/pkg/sqlformat"

	"uuid"

	"github.com/lib/pq"
)

type CategoryRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(
	db *sql.DB,
	logger *slog.Logger,
) *CategoryRepository {
	return &CategoryRepository{
		db:     db,
		logger: logger.WithGroup("db"),
	}
}

type repository interface {
	FindAll(
		ctx context.Context,
		search string,
		f filters.Filters,
	) ([]*Category, filters.Metadata, error)

	FindById(
		ctx context.Context,
		id uuid.UUID,
	) (*Category, error)

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

func parseConstraintError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Constraint {
		case "categories_user_id_name_type_key", "uniq_category_user_name_type":
			return apiError.ValidationAlreadyExists("name")
		}
	}
	return err
}

func scanType(s string) (CategoryType, error) {
	return ParseCategoryType(s)
}

func (r *CategoryRepository) FindAll(
	ctx context.Context,
	search string,
	f filters.Filters,
) ([]*Category, filters.Metadata, error) {
	userAuth := contexts.GetUser(ctx)

	query := fmt.Sprintf(`
        SELECT
            count(*) OVER(),
            id,
            user_id,
            name,
            type,
            goal_id,
            version,
            created_at,
            created_by,
            updated_at,
            updated_by
        FROM categories
        WHERE deleted = false
        AND (
            name ILIKE '%%' || $1 || '%%'
            OR $1 = ''
        )
		and user_id = $4
        ORDER BY %s %s, id ASC
        LIMIT $2 OFFSET $3
    `, f.SortColumn(), f.SortDirection())

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	args := []any{search, f.Limit(), f.Offset(), userAuth.GetID()}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, filters.Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	models := make([]*Category, 0)

	for rows.Next() {
		var model Category

		err := rows.Scan(
			&totalRecords,
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.Type,
			&model.GoalID,
			&model.Version,
			&model.CreatedAt,
			&model.CreatedBy,
			&model.UpdatedAt,
			&model.UpdatedBy,
		)
		if err != nil {
			return nil, filters.Metadata{}, err
		}

		models = append(models, &model)
	}

	if err = rows.Err(); err != nil {
		return nil, filters.Metadata{}, err
	}

	metadata := filters.CalculateMetadata(totalRecords, f.Page, f.PageSize)
	return models, metadata, nil
}

func (r *CategoryRepository) FindById(
	ctx context.Context,
	id uuid.UUID,
) (*Category, error) {
	userAuth := contexts.GetUser(ctx)

	query := `
		select
			id,
			user_id,
			name,
			type,
			goal_id,
			version,
			created_at,
			created_by,
			updated_at,
			updated_by
		from categories
		where
			deleted = false
			and id = $1
			and user_id = $2
	`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	var model Category

	err := r.db.QueryRowContext(ctx, query, id, userAuth.GetID()).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.Type,
		&model.GoalID,
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

func (r *CategoryRepository) Insert(
	ctx context.Context,
	model *Category,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		insert into categories (
			user_id,
			name,
			type,
			goal_id,
			created_by
		) values (
			:user_id,
			:name,
			:type,
			:goal_id,
			:created_by
		) returning id, created_at, version
	`

	params := map[string]any{
		"user_id":    userAuth.GetID(),
		"name":       model.Name,
		"type":       model.Type,
		"goal_id":    model.GoalID,
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
		return parseConstraintError(err)
	}

	return nil
}

func (r *CategoryRepository) Update(
	ctx context.Context,
	model *Category,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		update categories
		set
			name = :name,
			type = :type,
			goal_id = :goal_id,
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
		"name":    model.Name,
		"type":    model.Type,
		"goal_id": model.GoalID,
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

	err := tx.QueryRowContext(ctx, query, args...).Scan(&model.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apiError.ErrEditConflict
		}
		return parseConstraintError(err)
	}

	return nil
}

func (r *CategoryRepository) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		update categories
		set
			deleted = true,
			updated_at = NOW(),
			updated_by = :user_id,
			version = version + 1
		where
			id = :id
			and deleted = false
			and user_id = :user_id
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
