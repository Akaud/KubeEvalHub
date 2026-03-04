package config

import (
	"os"
	"strconv"
	"time"
)

func String(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func Int(key string, fallback int) int {
	if s, ok := os.LookupEnv(key); ok {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return fallback
}

func DurationSeconds(key string, fallbackSeconds int) time.Duration {
	return time.Duration(Int(key, fallbackSeconds)) * time.Second
}
