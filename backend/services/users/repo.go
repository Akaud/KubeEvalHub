package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Akaud/KubeEvalHub/types"
)

var ErrNotFound = errors.New("not found")

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, username, email, passwordHash string) (types.User, error) {
	const q = `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, password_hash, created_at, updated_at
	`
	var u types.User
	err := r.db.QueryRowContext(ctx, q, username, email, passwordHash).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *Repo) Update(ctx context.Context, id int64, username, email, passwordHash *string) (types.User, error) {
	const q = `
		UPDATE users
		SET
		  username = COALESCE($2, username),
		  email = COALESCE($3, email),
		  password_hash = COALESCE($4, password_hash)
		WHERE id = $1
		RETURNING id, username, email, password_hash, created_at, updated_at
	`
	var u types.User
	err := r.db.QueryRowContext(ctx, q, id, username, email, passwordHash).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return types.User{}, ErrNotFound
	}
	return u, err
}

func (r *Repo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM users WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
