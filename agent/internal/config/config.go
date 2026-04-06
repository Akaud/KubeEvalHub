package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	BackendURL      string
	AgentToken      string
	ScrapeInterval  time.Duration
	RequestTimeout  time.Duration
	InsecureSkipTLS bool
}

func Load() Config {
	return Config{
		BackendURL:      getEnv("BACKEND_URL", "http://backend:8080"),
		AgentToken:      getEnv("AGENT_TOKEN", ""),
		ScrapeInterval:  getEnvDuration("SCRAPE_INTERVAL", 30*time.Second),
		RequestTimeout:  getEnvDuration("REQUEST_TIMEOUT", 10*time.Second),
		InsecureSkipTLS: getEnvBool("INSECURE_SKIP_TLS", false),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
