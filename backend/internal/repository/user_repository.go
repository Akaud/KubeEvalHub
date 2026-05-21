package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func scanUser(row pgx.Row) (*model.User, error) {
	var user model.User

	err := row.Scan(
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
		return nil, err
	}

	return &user, nil
}

func scanUserWithVerification(row pgx.Row) (*model.User, error) {
	var user model.User
	var verifyTokenExpiry *time.Time

	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.EmailVerified,
		&user.EmailVerifyToken,
		&verifyTokenExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.VerifyTokenExpiry = verifyTokenExpiry
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, name, email, password string) (*model.User, error) {
	const query = `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password, created_at, updated_at
	`

	user, err := scanUser(r.pool.QueryRow(ctx, query, name, email, password))
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) CreateWithVerification(ctx context.Context, name, email, passwordHash, verifyToken string, tokenExpiry time.Time) (*model.User, error) {
	const query = `
		INSERT INTO users (name, email, password, email_verified, email_verify_token, verify_token_expiry, created_at, updated_at)
		VALUES ($1, $2, $3, false, $4, $5, NOW(), NOW())
		RETURNING id, name, email, password, email_verified, email_verify_token, verify_token_expiry, created_at, updated_at
	`

	user, err := scanUserWithVerification(r.pool.QueryRow(ctx, query, name, email, passwordHash, verifyToken, tokenExpiry))
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("create user with verification: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const query = `
		SELECT id, name, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmailVerified(ctx context.Context, email string) (*model.User, error) {
	const query = `
		SELECT id, name, email, password, email_verified, created_at, updated_at
		FROM users
		WHERE email = $1 AND email_verified = true
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email verified: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByEmailWithVerification(ctx context.Context, email string) (*model.User, error) {
	const query = `
		SELECT id, name, email, password, email_verified, email_verify_token, verify_token_expiry, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user, err := scanUserWithVerification(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email with verification: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByName(ctx context.Context, name string) (*model.User, error) {
	const query = `
		SELECT id, name, email, password, email_verified, created_at, updated_at
		FROM users
		WHERE name = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by name: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	const query = `
		SELECT id, name, email, password, email_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
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
		RETURNING id, name, email, password, email_verified, created_at, updated_at
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, id, name, email, password).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.EmailVerified,
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
		RETURNING id, name, email, password, email_verified, created_at, updated_at
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, id, name, email, password).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.EmailVerified,
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

func (r *UserRepository) VerifyEmail(ctx context.Context, token string) (*model.User, error) {
	const query = `
		UPDATE users 
		SET email_verified = true, 
		    email_verify_token = NULL, 
		    verify_token_expiry = NULL,
		    updated_at = NOW()
		WHERE email_verify_token = $1 
		  AND verify_token_expiry > NOW()
		  AND email_verified = false
		RETURNING id, name, email, password, email_verified, created_at, updated_at
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidVerificationToken
		}
		return nil, fmt.Errorf("verify email: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateVerificationToken(ctx context.Context, userID int64, token string, expiry time.Time) error {
	const query = `
		UPDATE users
		SET email_verify_token = $2,
		    verify_token_expiry = $3,
		    email_verified = false,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`

	var id int64
	err := r.pool.QueryRow(ctx, query, userID, token, expiry).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("update verification token: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUnverifiedUsers(ctx context.Context, olderThan time.Time) ([]*model.User, error) {
	const query = `
		SELECT id, name, email, password, email_verified, email_verify_token, verify_token_expiry, created_at, updated_at
		FROM users
		WHERE email_verified = false 
		  AND verify_token_expiry < $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, olderThan)
	if err != nil {
		return nil, fmt.Errorf("get unverified users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		var verifyTokenExpiry *time.Time
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.EmailVerified,
			&user.EmailVerifyToken,
			&verifyTokenExpiry,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan unverified user: %w", err)
		}
		user.VerifyTokenExpiry = verifyTokenExpiry
		users = append(users, &user)
	}

	return users, nil
}

func (r *UserRepository) DeleteUnverifiedUsers(ctx context.Context, olderThan time.Time) (int64, error) {
	const query = `
		DELETE FROM users
		WHERE email_verified = false 
		  AND verify_token_expiry < $1
	`

	tag, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("delete unverified users: %w", err)
	}

	return tag.RowsAffected(), nil
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
