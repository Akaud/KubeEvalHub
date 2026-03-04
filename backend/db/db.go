package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string

	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetimeSec int
	PingTimeoutSec     int
}

func Open(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetimeSec > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSec) * time.Second)
	}

	return sqlDB, nil
}

func Ping(ctx context.Context, db *sql.DB, timeoutSec int) error {
	if timeoutSec <= 0 {
		timeoutSec = 3
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}
