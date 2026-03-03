package users

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Akaud/KubeEvalHub/types"
)

// ErrNotFound is returned when a user record does not exist.
var ErrNotFound = errors.New("not found")

// Repo provides access to user persistence operations.
type Repo struct {
	db *sql.DB
}

// NewRepo creates a new user repository backed by sql.DB.
func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// Create inserts a new user record and returns the persisted user.
// The password must already be hashed before calling this method.
func (r *Repo) Create(
	ctx context.Context,
	username, email, passwordHash string,
) (types.User, error) {

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

// Update modifies an existing user by ID.
// Only non-nil fields are updated.
// Returns ErrNotFound if the user does not exist.
func (r *Repo) Update(
	ctx context.Context,
	id int64,
	username, email, passwordHash *string,
) (types.User, error) {

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

// Delete removes a user by ID.
// Returns ErrNotFound if the user does not exist.
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

// GetByID retrieves a user by primary key.
// Returns ErrNotFound if no record exists.
func (r *Repo) GetByID(ctx context.Context, id int64) (types.User, error) {
	const q = `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var u types.User
	err := r.db.QueryRowContext(ctx, q, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return types.User{}, ErrNotFound
	}

	return u, err
}

// GetByLogin retrieves a user by email or username (case-insensitive).
// Used for authentication lookups.
// Returns ErrNotFound if no record matches.
func (r *Repo) GetByLogin(ctx context.Context, login string) (types.User, error) {
	login = strings.TrimSpace(strings.ToLower(login))

	const q = `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE lower(email) = $1 OR lower(username) = $1
		LIMIT 1
	`

	var u types.User
	err := r.db.QueryRowContext(ctx, q, login).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return types.User{}, ErrNotFound
	}

	return u, err
}
