package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	projectmigrations "backend/migrations"
)

func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if err := applyMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	goose.SetBaseFS(projectmigrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() {
		_ = sqlDB.Close()
	}()

	if err := pingSQLDB(ctx, sqlDB); err != nil {
		return fmt.Errorf("ping sql db for migrations: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, "."); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func pingSQLDB(ctx context.Context, db *sql.DB) error {
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
