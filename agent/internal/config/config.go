package config

import (
	"fmt"
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

func Load() (Config, error) {
	var cfg Config
	var err error

	if cfg.BackendURL, err = requireEnv("BACKEND_URL"); err != nil {
		return Config{}, err
	}

	if cfg.AgentToken, err = requireEnv("AGENT_TOKEN"); err != nil {
		return Config{}, err
	}

	if cfg.ScrapeInterval, err = requireEnvDuration("SCRAPE_INTERVAL"); err != nil {
		return Config{}, err
	}

	if cfg.RequestTimeout, err = requireEnvDuration("REQUEST_TIMEOUT"); err != nil {
		return Config{}, err
	}

	if cfg.InsecureSkipTLS, err = requireEnvBool("INSECURE_SKIP_TLS"); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("missing required env: %s", key)
	}
	return v, nil
}

func requireEnvDuration(key string) (time.Duration, error) {
	v, err := requireEnv(key)
	if err != nil {
		return 0, err
	}
	return time.ParseDuration(v)
}

func requireEnvBool(key string) (bool, error) {
	v, err := requireEnv(key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(v)
}
