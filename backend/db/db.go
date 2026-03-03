package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
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
}

func LoadConfigFromEnv() Config {
	return Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Name:     getEnv("DB_NAME", "postgres"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
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

	sqlDB.SetMaxOpenConns(getEnvAsInt("DB_MAX_OPEN_CONNS", 10))
	sqlDB.SetMaxIdleConns(getEnvAsInt("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetConnMaxLifetime(
		time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME_SEC", 300)) * time.Second,
	)

	return sqlDB, nil
}

func Ping(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(getEnvAsInt("DB_PING_TIMEOUT_SEC", 3))*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

func Close(db *sql.DB) error {
	return db.Close()
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if s, ok := os.LookupEnv(key); ok {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return fallback
}
