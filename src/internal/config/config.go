package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort       string
	JWTSigningKey    string
	JWTTokenDuration string
	DatabaseURL      string
	MetricsPort      string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnvVar("POSTGRES_USER", "myuser"),
		getEnvVar("POSTGRES_PASSWORD", "wasted"),
		getEnvVar("POSTGRES_HOST", "localhost"),
		getEnvVar("POSTGRES_PORT", "5432"),
		getEnvVar("POSTGRES_DB", "Avito-backend"),
	)

	return &Config{
		ServerPort:       getEnvVar("SERVER_PORT", "8080"),
		JWTSigningKey:    getEnvVar("JWT_SIGNING_KEY", "default-secret-key"),
		JWTTokenDuration: getEnvVar("JWT_TOKEN_DURATION", "24h"),
		MetricsPort:      getEnvVar("METRICS_PORT", "9000"),
		DatabaseURL:      dbURL,
	}, nil
}

func getEnvVar(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
