package transactions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"ms_transaction/internal/core/contexts"
	"ms_transaction/internal/core/domain/apiError"
	"ms_transaction/internal/core/filters"
	"ms_transaction/pkg/sqlformat"
	"time"
	"uuid"

	"github.com/lib/pq"
)

type TransactionRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(
	db *sql.DB,
	logger *slog.Logger,
) *TransactionRepository {
	return &TransactionRepository{
		db:     db,
		logger: logger.WithGroup("db"),
	}
}

type transactionQuery struct {
	StartDate  *time.Time
	EndDate    *time.Time
	Type       *CategoryType
	CategoryID uuid.UUID
	Filters    filters.Filters
}

func parseConstraintError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Constraint {
		case "transactions_user_id_category_id_key":
			return apiError.ValidationAlreadyExists("category_id")
		}
	}
	return err
}

type repository interface {
	FindAll(
		ctx context.Context,
		query transactionQuery,
	) ([]*Transaction, filters.Metadata, error)

	FindById(
		ctx context.Context,
		id uuid.UUID,
	) (*Transaction, error)

	Insert(
		ctx context.Context,
		model *Transaction,
	) error

	Update(
		ctx context.Context,
		model *Transaction,
	) error

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) error
}

func (r *TransactionRepository) FindAll(
	ctx context.Context,
	query transactionQuery,
) ([]*Transaction, filters.Metadata, error) {
	userAuth := contexts.GetUser(ctx)

	q := fmt.Sprintf(`
	SELECT
        count(*) OVER(),
        t.id,
        t.user_id,
        t.amount,
        t.category_id,
        t.description,
        t.version,
        t.created_at,
        t.created_by,
        t.updated_at,
        t.updated_by
    FROM transactions t
	left join categories c
		on c.id = t.category_id
	where deleted = false
        AND user_id = :userID
        AND (:startDate::timestamptz IS NULL OR t.created_at >= :startDate)
        AND (:endDate::timestamptz IS NULL OR t.created_at <= :endDate)
		AND (:categoryID::uuid IS NULL OR t.category_id = :categoryID)
		AND (:type::integer IS NULL OR c.type = :type)
    ORDER BY %s %s, id ASC
    LIMIT :limit OFFSET :offset
    `, query.Filters.SortColumn(), query.Filters.SortDirection())

	params := map[string]any{
		"userID":     userAuth.GetID(),
		"startDate":  query.StartDate,
		"endDate":    query.EndDate,
		"categoryID": query.CategoryID,
		"type":       query.Type,
		"limit":      query.Filters.Limit(),
		"offset":     query.Filters.Offset(),
	}

	q, args := sqlformat.NamedQuery(q, params)

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, filters.Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	transactions := make([]*Transaction, 0)

	for rows.Next() {
		var model Transaction
		err := rows.Scan(
			&totalRecords,
			&model.ID,
			&model.UserID,
			&model.Amount,
			&model.CategoryID,
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
		transactions = append(transactions, &model)
	}

	if err = rows.Err(); err != nil {
		return nil, filters.Metadata{}, err
	}

	metadata := filters.CalculateMetadata(totalRecords, query.Filters.Page, query.Filters.PageSize)
	return transactions, metadata, nil
}

func nullUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil() {
		return nil
	}
	return &id
}

func (r *TransactionRepository) FindById(
	ctx context.Context,
	id uuid.UUID,
) (*Transaction, error) {
	userAuth := contexts.GetUser(ctx)

	query := `
		SELECT
			id,
			user_id,
			amount,
			category_id,
			description,
			version,
			created_at,
			created_by,
			updated_at,
			updated_by
		FROM transactions
		WHERE
			deleted = false
			AND id = $1
			AND user_id = $2
	`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	var model Transaction
	err := r.db.QueryRowContext(ctx, query, id, userAuth.GetID()).Scan(
		&model.ID,
		&model.UserID,
		&model.Amount,
		&model.CategoryID,
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

	return &model, nil
}

func (r *TransactionRepository) Insert(
	ctx context.Context,
	model *Transaction,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		INSERT INTO transactions (
			user_id,
			amount,
			category_id,
			description,
			created_by
		) VALUES (
			:user_id,
			:amount,
			:category_id,
			:description,
			:created_by
		) RETURNING id, created_at, version
	`

	params := map[string]any{
		"user_id":     userAuth.GetID(),
		"amount":      model.Amount,
		"category_id": model.CategoryID,
		"description": model.Description,
		"created_by":  userAuth.GetID(),
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

func (r *TransactionRepository) Update(
	ctx context.Context,
	model *Transaction,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		UPDATE transactions
		SET
			amount = :amount,
			category_id = :category_id,
			description = :description,
			updated_at = NOW(),
			updated_by = :user_id,
			version = version + 1
		WHERE
			id = :id
			and user_id = :user_id
			AND version = :version
			AND deleted = false
		RETURNING version
	`

	params := map[string]any{
		"amount":      model.Amount,
		"category_id": model.CategoryID,
		"description": model.Description,
		"id":          model.ID,
		"version":     model.Version,
		"user_id":     userAuth.GetID(),
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

func (r *TransactionRepository) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
		UPDATE transactions
		SET
			deleted = true,
			updated_at = NOW(),
			updated_by = :user_id,
			version = version + 1
		WHERE
			id = :id
			AND deleted = false
			AND user_id = :user_id
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
