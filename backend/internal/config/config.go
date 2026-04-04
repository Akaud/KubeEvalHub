package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration.
// Single source of truth for environment-derived values.
type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
}

// Load reads environment variables and validates them.
func Load() *Config {
	cfg := &Config{
		Port:         mustGetEnv("PORT"),
		ReadTimeout:  mustGetDuration("READ_TIMEOUT"),
		WriteTimeout: mustGetDuration("WRITE_TIMEOUT"),
		IdleTimeout:  mustGetDuration("IDLE_TIMEOUT"),

		DBHost:     mustGetEnv("DB_HOST"),
		DBPort:     mustGetEnv("DB_PORT"),
		DBUser:     mustGetEnv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"), // optional
		DBName:     mustGetEnv("DB_NAME"),
		DBSSLMode:  mustGetEnv("DB_SSLMODE"),
		JWTSecret:  mustGetEnv("JWT_SECRET"),
	}

	return cfg
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return v
}

// Supports values like: "5" (seconds) or "5s"
func mustGetDuration(key string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("%s is required", key))
	}

	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return d
	}

	if i, err := strconv.Atoi(v); err == nil && i > 0 {
		return time.Duration(i) * time.Second
	}

	panic(fmt.Sprintf("%s must be a valid positive duration", key))
}

func (c *Config) DatabaseURL() string {
	auth := c.DBUser
	if c.DBPassword != "" {
		auth += ":" + c.DBPassword
	}

	return fmt.Sprintf(
		"postgres://%s@%s:%s/%s?sslmode=%s",
		auth,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}
