package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	Port              string
	Redis             RedisConfig
	APIKey            string
	StorageDriverName string
	Storage           StorageConfig
	Auth              AuthConfig
}

type StorageConfig struct {
	Driver   string
	Postgres PostgresConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

type RedisConfig struct {
	Host string
	Port string
}

type AuthConfig struct {
	CookieSecret string
}

func Load() (*Config, error) {
	// Load .env if it exists.
	// In production, environment variables will already be provided.
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Port:   getEnv("PORT", "4215"),

		Storage: StorageConfig{
			Driver: getEnv("STORAGE_DRIVER", "sqlite"),
		},

		Redis: RedisConfig{
			Host: getRequiredEnv("REDIS_HOST"),
			Port: getEnv("REDIS_PORT", "6379"),
		},

		Auth: AuthConfig{
			CookieSecret: getRequiredEnv("RATEGUARD_COOKIE_SECRET"),
		},
	}

	if cfg.Storage.Driver == "postgres" {
		cfg.Storage.Postgres = PostgresConfig{
			Host:     getRequiredEnv("POSTGRES_HOST"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			Database: getRequiredEnv("POSTGRES_DATABASE"),
			User:     getRequiredEnv("POSTGRES_USER"),
			Password: getRequiredEnv("POSTGRES_PASSWORD"),
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf(
			"required environment variable %s is not set",
			key,
		))
	}

	return value
}

func (c *Config) StorageDriver() string {
	if c.Storage.Driver == "" {
		return "sqlite"
	}

	return c.Storage.Driver
}
