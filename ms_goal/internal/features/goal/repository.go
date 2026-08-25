package goal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"ms_goal/internal/core/contexts"
	"ms_goal/internal/core/domain/apiError"
	"ms_goal/internal/core/filters"
	"ms_goal/pkg/sqlformat"
	"uuid"

	"github.com/lib/pq"
)

type GoalRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(
	db *sql.DB,
	logger *slog.Logger,
) *GoalRepository {
	return &GoalRepository{
		db:     db,
		logger: logger.WithGroup("db"),
	}
}

func parseConstraintError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Constraint {
		case "goal_user_id_name_type_key":
			return apiError.ValidationAlreadyExists("name")
		}
	}
	return err
}

func scanType(s string) (GoalStatus, error) {
	return ParseGoalStatus(s)
}

type repository interface {
	FindAll(ctx context.Context, search string, status GoalStatus, f filters.Filters) ([]*Goal, filters.Metadata, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Goal, error)
	Insert(ctx context.Context, model *Goal) error
	Update(ctx context.Context, model *Goal) error
	DeleteById(ctx context.Context, id uuid.UUID) error
}

func (r *GoalRepository) FindAll(
	ctx context.Context,
	search string,
	status GoalStatus,
	f filters.Filters,
) ([]*Goal, filters.Metadata, error) {
	userAuth := contexts.GetUser(ctx)

	q := fmt.Sprintf(`
	select
		count(*) over(),
		g.id,
		g.user_id,
		g.name,
		g.target_amount,
		g.current_amount,
		g.status,
		g.deadline,
		g.description,
		g.version,
        g.created_at,
        g.created_by,
        g.updated_at,
        g.updated_by
	from goals g
	where deleted = false
		and user_id = :userID
		and (:status::integer is null or g.status = :status)
		and (
			name ilike '%%' || :search || '%%'
			or :search = ''
		)
		order by %s %s, id asc
		limit :limit offset :offset
	`, f.SortColumn(), f.SortDirection())

	params := map[string]any{
		"userID": userAuth.GetID(),
		"status": status,
		"search": search,
		"limit":  f.Limit(),
		"offset": f.Offset(),
	}

	query, args := sqlformat.NamedQuery(q, params)
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, filters.Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	models := make([]*Goal, 0)

	for rows.Next() {
		var model Goal

		err := rows.Scan(
			&totalRecords,
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.TargetAmount,
			&model.CurrentAmount,
			&model.Status,
			&model.Deadline,
			&model.Description,
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

func (r *GoalRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*Goal, error) {
	userAuth := contexts.GetUser(ctx)

	q := `
	select
		g.id,
		g.user_id,
		g.name,
		g.target_amount,
		g.current_amount,
		g.status,
		g.deadline,
		g.description,
		g.version,
        g.created_at,
        g.created_by,
        g.updated_at,
        g.updated_by
	from goals g
	where deleted = false
		and user_id = $1
		and id = $2
	`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	var model Goal

	err := r.db.QueryRowContext(ctx, q, userAuth.GetID(), id).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.TargetAmount,
		&model.CurrentAmount,
		&model.Status,
		&model.Deadline,
		&model.Description,
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

	return &model, err
}

func (r *GoalRepository) Insert(
	ctx context.Context,
	model *Goal,
) error {
	userAuth := contexts.GetUser(ctx)
	q := `
	insert into goals (
		user_id,
		name,
		target_amount,
		deadline,
		description,
		created_by
	) values (
 		:user_id,
		:name,
		:target_amount,
		:deadline,
		:description,
		:created_by
	) returning id, created_at, version
	`

	params := map[string]any{
		"user_id":       userAuth.GetID(),
		"name":          model.Name,
		"target_amount": model.TargetAmount,
		"deadline":      model.Deadline,
		"description":   model.Description,
		"created_by":    userAuth.GetID(),
	}

	query, args := sqlformat.NamedQuery(q, params)
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

func (r *GoalRepository) Update(
	ctx context.Context,
	model *Goal,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		update goals
		set
			name = :name,
			target_amount = :target_amount,
			current_amount = :current_amount,
			status = :status,
			deadline = :deadline,
			description = :description,
			updated_at = NOW(),
			updated_by = :user_id,
			version = version + 1
		where
			id = :id
			and user_id = :user_id
			and version = :version
			and deleted = false
		returning version
	`

	params := map[string]any{
		"name":           model.Name,
		"target_amount":  model.TargetAmount,
		"current_amount": model.CurrentAmount,
		"status":         model.Status,
		"deadline":       model.Deadline,
		"description":    model.Description,
		"id":             model.ID,
		"user_id":        userAuth.GetID(),
		"version":        model.Version,
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

func (r *GoalRepository) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		update goals
		set
			deleted = true,
			updated_at = NOW(),
			updated_by = :user_id,
			version = version + 1
		where
			id = :id
			and user_id = :user_id
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
