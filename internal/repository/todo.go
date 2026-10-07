package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/encourage13/todo-api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepository struct {
	pool *pgxpool.Pool
}

func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{pool: pool}
}

func (r *TodoRepository) Create(ctx context.Context, t *model.Todo) error {
	query := `INSERT INTO todos(user_id, title, description) VALUES ($1, $2, $3) RETURNING id, completed, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, t.UserID, t.Title, t.Description).Scan(&t.ID, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	return nil
}

func (r *TodoRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Todo, error) {
	var t model.Todo
	query := `SELECT id, user_id, title, completed, description, created_at, updated_at FROM todos WHERE id=$1 AND user_id=$2`

	err := r.pool.QueryRow(ctx, query, id, userID).Scan(&t.ID, &t.UserID, &t.Title, &t.Completed, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTodoNotFound
		}
		return nil, fmt.Errorf("task by id: %w", err)
	}

	return &t, nil
}

func (r *TodoRepository) List(ctx context.Context, userID uuid.UUID) ([]model.Todo, error) {
	query := `SELECT id, user_id, title, completed, description, created_at, updated_at FROM todos WHERE user_id=$1
			  ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get all tasks by user_id: %w", err)
	}
	defer rows.Close()

	tasks := make([]model.Todo, 0)
	for rows.Next() {
		var t model.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Completed, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}
