package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL      string
	DatabaseMaxConns int

	// Redis
	RedisURL      string
	RedisPassword string

	// Kafka
	KafkaBrokers []string

	// JWT
	JWTSecret       string
	JWTExpiry       string
	JWTRefreshExpiry string

	// Tracing
	JaegerEndpoint string

	// External APIs
	CreditBureauAPIURL string
	CreditBureauAPIKey string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                getEnv("PORT", "8001"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		DatabaseMaxConns:    getEnvAsInt("DATABASE_MAX_CONNECTIONS", 50),
		RedisURL:            getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		KafkaBrokers:        []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTExpiry:           getEnv("JWT_EXPIRY", "15m"),
		JWTRefreshExpiry:    getEnv("JWT_REFRESH_EXPIRY", "7d"),
		JaegerEndpoint:      getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		CreditBureauAPIURL:  getEnv("CREDIT_BUREAU_API_URL", ""),
		CreditBureauAPIKey:  getEnv("CREDIT_BUREAU_API_KEY", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
