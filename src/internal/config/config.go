package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	JWTSigningKey string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		ServerPort:    getEnvVar("SERVER_PORT", "8080"),
		JWTSigningKey: getEnvVar("JWT_SIGNING_KEY", "default-secret-key"),
	}, nil
}

func getEnvVar(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
