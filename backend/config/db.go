package config

import "github.com/Akaud/KubeEvalHub/db"

func DB() db.Config {
	return db.Config{
		Host:     String("DB_HOST", "localhost"),
		Port:     Int("DB_PORT", 5432),
		User:     String("DB_USER", "postgres"),
		Password: String("DB_PASSWORD", "postgres"),
		Name:     String("DB_NAME", "postgres"),
		SSLMode:  String("DB_SSLMODE", "disable"),

		MaxOpenConns:       Int("DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns:       Int("DB_MAX_IDLE_CONNS", 10),
		ConnMaxLifetimeSec: Int("DB_CONN_MAX_LIFETIME_SEC", 300),
		PingTimeoutSec:     Int("DB_PING_TIMEOUT_SEC", 3),
	}
}
