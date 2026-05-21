-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL, -- hashed password only
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verify_token TEXT,
    verify_token_expiry TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index for faster token lookups during verification
CREATE INDEX idx_users_email_verify_token ON users(email_verify_token);

-- Create index for finding expired unverified users (useful for cleanup jobs)
CREATE INDEX idx_users_unverified_expired ON users(email_verified, verify_token_expiry) 
    WHERE email_verified = FALSE AND verify_token_expiry IS NOT NULL;

-- Optional: Create a partial unique index to ensure only one active verification token per email
-- This prevents multiple verification tokens for the same unverified email
CREATE UNIQUE INDEX idx_users_active_verification ON users(email) 
    WHERE email_verified = FALSE AND email_verify_token IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_active_verification;
DROP INDEX IF EXISTS idx_users_unverified_expired;
DROP INDEX IF EXISTS idx_users_email_verify_token;
DROP TABLE IF EXISTS users;