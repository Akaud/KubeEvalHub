package repository

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) Create(ctx context.Context, name, email, password string) (*model.User, error) {
	const query = `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password, created_at, updated_at
	`

	var user model.User

	err := r.pool.QueryRow(ctx, query, name, email, password).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, id int64, name, email, password string) (*model.User, error) {
	const query = `
		UPDATE users
		SET name = $2,
		    email = $3,
		    password = $4,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, password, created_at, updated_at
	`

	var user model.User

	err := r.pool.QueryRow(ctx, query, id, name, email, password).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Patch(ctx context.Context, id int64, name, email, password *string) (*model.User, error) {
	const query = `
		UPDATE users
		SET name = COALESCE($2, name),
		    email = COALESCE($3, email),
		    password = COALESCE($4, password),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, password, created_at, updated_at
	`

	var user model.User

	err := r.pool.QueryRow(ctx, query, id, name, email, password).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("patch user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM users WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
