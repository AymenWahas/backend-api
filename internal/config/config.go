package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	RequestTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port: getEnv("PORT", "8443"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5434"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgrespassword"),
		DBName:     getEnv("DB_NAME", "employee_db"),

		RequestTimeout: getEnvDuration(
			"REQUEST_TIMEOUT",
			5*time.Second,
		),

		ConnMaxLifetime: getEnvDuration(
			"DB_CONN_MAX_LIFETIME",
			30*time.Minute,
		),

		ConnMaxIdleTime: getEnvDuration(
			"DB_CONN_MAX_IDLE_TIME",
			5*time.Minute,
		),
	}

	var err error

	cfg.MaxOpenConns, err = getEnvInt(
		"DB_MAX_OPEN_CONNS",
		25,
	)
	if err != nil {
		return Config{}, err
	}

	cfg.MaxIdleConns, err = getEnvInt(
		"DB_MAX_IDLE_CONNS",
		10,
	)
	if err != nil {
		return Config{}, err
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	if cfg.RequestTimeout <= 0 {
		return fmt.Errorf("REQUEST_TIMEOUT must be greater than zero")
	}

	if cfg.MaxOpenConns <= 0 {
		return fmt.Errorf(
			"DB_MAX_OPEN_CONNS must be greater than zero",
		)
	}

	if cfg.MaxIdleConns < 0 {
		return fmt.Errorf(
			"DB_MAX_IDLE_CONNS cannot be negative",
		)
	}

	if cfg.MaxIdleConns > cfg.MaxOpenConns {
		return fmt.Errorf(
			"DB_MAX_IDLE_CONNS cannot be greater than DB_MAX_OPEN_CONNS",
		)
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(
	key string,
	fallback int,
) (int, error) {
	value := os.Getenv(key)

	if value == "" {
		return fallback, nil
	}

	result, err := strconv.Atoi(value)

	if err != nil {
		return 0, fmt.Errorf(
			"%s must be an integer: %w",
			key,
			err,
		)
	}

	return result, nil
}

func getEnvDuration(
	key string,
	fallback time.Duration,
) time.Duration {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	result, err := time.ParseDuration(value)

	if err != nil {
		return fallback
	}

	return result
}
