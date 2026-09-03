package goaltransaction

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

type GoalTransactionRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(
	db *sql.DB,
	logger *slog.Logger,
) *GoalTransactionRepository {
	return &GoalTransactionRepository{
		db:     db,
		logger: logger.WithGroup("db"),
	}
}

type repository interface {
	FindAll(ctx context.Context, goalID uuid.UUID, f filters.Filters) ([]*GoalTransaction, filters.Metadata, error)
	FindByID(ctx context.Context, id uuid.UUID) (*GoalTransaction, error)
	Insert(ctx context.Context, model *GoalTransaction) error
	Update(
		ctx context.Context,
		model *GoalTransaction,
	) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	DeleteByGoalId(
		ctx context.Context,
		goalID uuid.UUID,
	) error
}

func parseConstraintError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Constraint {
		case "goal_transactions_goal_id_transaction_id_key":
			return apiError.ValidationAlreadyExists("transaction_id")
		}
	}
	return err
}

func (r *GoalTransactionRepository) FindAll(
	ctx context.Context,
	goalID uuid.UUID,
	f filters.Filters,
) ([]*GoalTransaction, filters.Metadata, error) {
	userAuth := contexts.GetUser(ctx)

	q := fmt.Sprintf(`
        SELECT
            count(*) OVER(),
            gt.id,
            gt.user_id,
            gt.goal_id,
            gt.transaction_id,
            gt.amount,
            gt.version,
            gt.created_at,
            gt.created_by,
            gt.updated_at,
            gt.updated_by
        FROM goal_transactions gt
        WHERE gt.deleted = false
          AND gt.user_id = :userID
          AND (:goalID::uuid IS NULL OR gt.goal_id = :goalID)
        ORDER BY %s %s, gt.id ASC
        LIMIT :limit OFFSET :offset
    `, f.SortColumn(), f.SortDirection())

	params := map[string]any{
		"userID": userAuth.GetID(),
		"goalID": nullUUID(goalID),
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
	models := make([]*GoalTransaction, 0)

	for rows.Next() {
		var model GoalTransaction
		err := rows.Scan(
			&totalRecords,
			&model.ID,
			&model.UserID,
			&model.GoalID,
			&model.TransactionID,
			&model.Amount,
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

func (r *GoalTransactionRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*GoalTransaction, error) {
	userAuth := contexts.GetUser(ctx)

	q := `
        SELECT
            gt.id,
            gt.user_id,
            gt.goal_id,
            gt.transaction_id,
            gt.amount,
            gt.version,
            gt.created_at,
            gt.created_by,
            gt.updated_at,
            gt.updated_by
        FROM goal_transactions gt
        WHERE gt.deleted = false
          AND gt.id = $1
          AND gt.user_id = $2
    `

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	var model GoalTransaction
	err := r.db.QueryRowContext(ctx, q, id, userAuth.GetID()).Scan(
		&model.ID,
		&model.UserID,
		&model.GoalID,
		&model.TransactionID,
		&model.Amount,
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

func (r *GoalTransactionRepository) Insert(
	ctx context.Context,
	model *GoalTransaction,
) error {
	userAuth := contexts.GetUser(ctx)

	q := `
        INSERT INTO goal_transactions (
            user_id,
            goal_id,
            transaction_id,
            amount,
            created_by
        ) VALUES (
            :user_id,
            :goal_id,
            :transaction_id,
            :amount,
            :created_by
        ) RETURNING id, created_at, version
    `

	params := map[string]any{
		"user_id":        userAuth.GetID(),
		"goal_id":        model.GoalID,
		"transaction_id": model.TransactionID,
		"amount":         model.Amount,
		"created_by":     userAuth.GetID(),
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

func (r *GoalTransactionRepository) Update(
	ctx context.Context,
	model *GoalTransaction,
) error {
	userAuth := contexts.GetUser(ctx)

	q := `
        UPDATE goal_transactions
        SET
            amount = :amount,
            updated_at = NOW(),
            updated_by = :user_id,
            version = version + 1
        WHERE
            id = :id
            AND user_id = :user_id
            AND deleted = false
    `

	params := map[string]any{
		"id":      model.ID,
		"amount":  model.Amount,
		"user_id": userAuth.GetID(),
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

func (r *GoalTransactionRepository) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	q := `
        UPDATE goal_transactions
        SET
            deleted = true,
            updated_at = NOW(),
            updated_by = :user_id,
            version = version + 1
        WHERE
            id = :id
            AND user_id = :user_id
            AND deleted = false
    `

	params := map[string]any{
		"id":      id,
		"user_id": userAuth.GetID(),
	}

	query, args := sqlformat.NamedQuery(q, params)
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

func (r *GoalTransactionRepository) DeleteByGoalId(
	ctx context.Context,
	goalID uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	q := `
        UPDATE goal_transactions
        SET
            deleted = true,
            updated_at = NOW(),
            updated_by = :user_id,
            version = version + 1
        WHERE
            goal_id = :goalID
            AND user_id = :user_id
            AND deleted = false
    `

	params := map[string]any{
		"goalID":  goalID,
		"user_id": userAuth.GetID(),
	}

	query, args := sqlformat.NamedQuery(q, params)
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

func nullUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil() {
		return nil
	}
	return &id
}
