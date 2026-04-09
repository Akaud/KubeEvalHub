package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshToken struct {
	ID                  int64
	UserID              int64
	TokenHash           string
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	ReplacedByTokenHash *string
	CreatedAt           time.Time
}

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		pool: pool,
	}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token RefreshToken) error {
	const query = `
		INSERT INTO user_refresh_tokens (
			user_id,
			token_hash,
			expires_at,
			revoked_at,
			replaced_by_token_hash
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.RevokedAt,
		token.ReplacedByTokenHash,
	)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, replaced_by_token_hash, created_at
		FROM user_refresh_tokens
		WHERE token_hash = $1
	`

	var token RefreshToken

	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.ReplacedByTokenHash,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}

	return &token, nil
}

func (r *RefreshTokenRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	const query = `
		UPDATE user_refresh_tokens
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE token_hash = $1
	`

	tag, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token by hash: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, oldHash string, newToken RefreshToken) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin rotate refresh token tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	now := time.Now().UTC()

	const revokeOldQuery = `
		UPDATE user_refresh_tokens
		SET revoked_at = $2,
		    replaced_by_token_hash = $3
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2
	`

	tag, err := tx.Exec(ctx, revokeOldQuery, oldHash, now, newToken.TokenHash)
	if err != nil {
		return fmt.Errorf("revoke old refresh token: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrRefreshTokenNotFound
	}

	const insertNewQuery = `
		INSERT INTO user_refresh_tokens (
			user_id,
			token_hash,
			expires_at,
			revoked_at,
			replaced_by_token_hash
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = tx.Exec(
		ctx,
		insertNewQuery,
		newToken.UserID,
		newToken.TokenHash,
		newToken.ExpiresAt,
		newToken.RevokedAt,
		newToken.ReplacedByTokenHash,
	)
	if err != nil {
		return fmt.Errorf("insert new refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit rotate refresh token tx: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteExpiredOrRevoked(ctx context.Context) error {
	const query = `
		DELETE FROM user_refresh_tokens
		WHERE expires_at < NOW()
		   OR revoked_at IS NOT NULL
	`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("delete expired or revoked refresh tokens: %w", err)
	}

	return nil
}
