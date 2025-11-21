package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func (r *GenericRepository[I, T]) Find(ctx context.Context, filters []Filter, opts *QueryOptions) ([]*T, error) {
	query, args := r.buildSelectQuery(filters, opts)

	var entities []T
	err := r.db.SelectContext(ctx, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	result := make([]*T, len(entities))
	for i := range entities {
		result[i] = &entities[i]
	}

	return result, nil
}

func (r *GenericRepository[I, T]) FindOne(ctx context.Context, filters []Filter) (*T, error) {
	opts := &QueryOptions{Limit: 1}
	query, args := r.buildSelectQuery(filters, opts)

	var entity T
	err := r.db.GetContext(ctx, &entity, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &entity, nil
}

func (r *GenericRepository[I, T]) FindByID(ctx context.Context, id string) (*T, error) {
	filters := []Filter{{Field: "id", Operator: "=", Value: id}}
	return r.FindOne(ctx, filters)
}

func (r *GenericRepository[I, T]) Create(ctx context.Context, entity T) (*T, error) {
	fields, values, placeholders := r.buildInsertData(entity)

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		pq.QuoteIdentifier(r.tableName),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	var createdEntity T
	err := r.db.GetContext(ctx, &createdEntity, query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return &createdEntity, nil
}

func (r *GenericRepository[I, T]) Update(ctx context.Context, id string, entity T) (*T, error) {
	fields, values := r.buildUpdateData(entity)
	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	setClause := strings.Join(fields, ", ")
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d RETURNING *",
		pq.QuoteIdentifier(r.tableName),
		setClause,
		len(values)+1,
	)

	values = append(values, id)
	var updatedEntity T
	err := r.db.GetContext(ctx, &updatedEntity, query, values...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entity with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	return &updatedEntity, nil
}

func (r *GenericRepository[I, T]) Upsert(ctx context.Context, entity T, conflictColumns []string) (*T, error) {
	fields, values, placeholders := r.buildInsertData(entity)
	updateFields, _ := r.buildUpdateData(entity)

	if len(conflictColumns) == 0 {
		conflictColumns = []string{"id"}
	}

	conflictClause := strings.Join(conflictColumns, ", ")
	updateClause := strings.Join(updateFields, ", ")

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING *",
		pq.QuoteIdentifier(r.tableName),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
		conflictClause,
		updateClause,
	)

	var upsertedEntity T
	err := r.db.GetContext(ctx, &upsertedEntity, query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert entity: %w", err)
	}

	return &upsertedEntity, nil
}

func (r *GenericRepository[I, T]) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", pq.QuoteIdentifier(r.tableName))
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("entity with id %s not found", id)
	}

	return nil
}

func (r *GenericRepository[I, T]) DeleteMany(ctx context.Context, filters []Filter) (int64, error) {
	whereClause, args := r.buildWhereClause(filters)
	query := fmt.Sprintf("DELETE FROM %s%s", pq.QuoteIdentifier(r.tableName), whereClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete entities: %w", err)
	}

	return result.RowsAffected()
}

func (r *GenericRepository[I, T]) Count(ctx context.Context, filters []Filter) (int64, error) {
	whereClause, args := r.buildWhereClause(filters)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", pq.QuoteIdentifier(r.tableName), whereClause)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}

	return count, nil
}

func (r *GenericRepository[I, T]) Exists(ctx context.Context, filters []Filter) (bool, error) {
	count, err := r.Count(ctx, filters)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GenericRepository[I, T]) ExistsWithID(ctx context.Context, id string) (bool, error) {
	filters := []Filter{{Field: "id", Operator: "=", Value: id}}
	return r.Exists(ctx, filters)
}

func (r *GenericRepository[I, T]) FindWithPagination(ctx context.Context, filters []Filter, page, pageSize int, orderBy string) ([]*T, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalCount, err := r.Count(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	opts := &QueryOptions{
		OrderBy: orderBy,
		Limit:   pageSize,
		Offset:  (page - 1) * pageSize,
	}

	entities, err := r.Find(ctx, filters, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get paginated results: %w", err)
	}

	return entities, totalCount, nil
}

func (r *GenericRepository[I, T]) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %w", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
