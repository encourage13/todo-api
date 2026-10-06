package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/encourage13/todo-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

const (
	pgUniqueViolation = "23505"
)

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	query := `INSERT INTO users(password_hash, email) VALUES ($1, $2) RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query, u.PasswordHash, u.Email).Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "users_email_key" && pgErr.Code == pgUniqueViolation {
			return model.ErrEmailTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	query := `SELECT id, password_hash, email, created_at FROM users WHERE email=$1`

	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.PasswordHash, &u.Email, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}
