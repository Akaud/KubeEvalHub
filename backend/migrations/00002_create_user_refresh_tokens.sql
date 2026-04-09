-- +goose Up
CREATE TABLE user_refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    replaced_by_token_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_refresh_tokens_user_id
    ON user_refresh_tokens (user_id);

CREATE INDEX idx_user_refresh_tokens_expires_at
    ON user_refresh_tokens (expires_at);

CREATE INDEX idx_user_refresh_tokens_active_user_id
    ON user_refresh_tokens (user_id)
    WHERE revoked_at IS NULL;


-- +goose Down
DROP TABLE IF EXISTS user_refresh_tokens;