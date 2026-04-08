package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	BackendURL     string
	AgentToken     string
	ScrapeInterval time.Duration
	RequestTimeout time.Duration
}

const (
	minScrapeInterval = 5 * time.Second
	timeoutBuffer     = 1 * time.Second
)

func Load() (Config, error) {
	var cfg Config
	var err error

	if cfg.BackendURL, err = getRequired("BACKEND_URL"); err != nil {
		return Config{}, err
	}

	if cfg.AgentToken, err = getRequired("AGENT_TOKEN"); err != nil {
		return Config{}, err
	}

	if cfg.ScrapeInterval, err = getRequiredDuration("SCRAPE_INTERVAL"); err != nil {
		return Config{}, err
	}

	cfg.RequestTimeout = cfg.ScrapeInterval - timeoutBuffer

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	switch {
	case c.BackendURL == "":
		return fmt.Errorf("missing required env: BACKEND_URL")
	case c.AgentToken == "":
		return fmt.Errorf("missing required env: AGENT_TOKEN")
	case c.ScrapeInterval < minScrapeInterval:
		return fmt.Errorf("SCRAPE_INTERVAL must be >= %s", minScrapeInterval)
	case c.RequestTimeout <= 0:
		return fmt.Errorf("REQUEST_TIMEOUT must be > 0")
	case c.RequestTimeout >= c.ScrapeInterval:
		return fmt.Errorf("REQUEST_TIMEOUT must be < SCRAPE_INTERVAL")
	default:
		return nil
	}
}

func getRequired(key string) (string, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return "", fmt.Errorf("missing required env: %s", key)
	}
	return v, nil
}

func getRequiredDuration(key string) (time.Duration, error) {
	v, err := getRequired(key)
	if err != nil {
		return 0, err
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}

	return d, nil
}
