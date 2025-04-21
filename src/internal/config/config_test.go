package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvVar(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "Existing variable",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "Non-existent variable",
			key:          "NON_EXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "Empty variable",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			if tt.envValue != "" {
				err := os.Setenv(tt.key, tt.envValue)
				assert.NoError(t, err)
			}

			result := getEnvVar(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
		wantErr  bool
	}{
		{
			name:    "Default values",
			envVars: map[string]string{},
			expected: &Config{
				ServerPort:       "8080",
				JWTSigningKey:    "default-secret-key",
				JWTTokenDuration: "24h",
				MetricsPort:      "9000",
				DatabaseURL:      "postgres://myuser:wasted@localhost:5432/Avito-backend?sslmode=disable",
			},
			wantErr: false,
		},
		{
			name: "Custom values",
			envVars: map[string]string{
				"SERVER_PORT":        "3000",
				"JWT_SIGNING_KEY":    "custom-key",
				"JWT_TOKEN_DURATION": "12h",
				"METRICS_PORT":       "9090",
				"POSTGRES_USER":      "testuser",
				"POSTGRES_PASSWORD":  "testpass",
				"POSTGRES_HOST":      "testhost",
				"POSTGRES_PORT":      "5433",
				"POSTGRES_DB":        "testdb",
			},
			expected: &Config{
				ServerPort:       "3000",
				JWTSigningKey:    "custom-key",
				JWTTokenDuration: "12h",
				MetricsPort:      "9090",
				DatabaseURL:      "postgres://testuser:testpass@testhost:5433/testdb?sslmode=disable",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile, err := os.Create(".env")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				envFile.Close()
				os.Remove(".env")
			}()

			os.Clearenv()

			for key, value := range tt.envVars {
				err := os.Setenv(key, value)
				assert.NoError(t, err)
			}

			config, err := LoadConfig()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.expected.ServerPort, config.ServerPort)
				assert.Equal(t, tt.expected.JWTSigningKey, config.JWTSigningKey)
				assert.Equal(t, tt.expected.JWTTokenDuration, config.JWTTokenDuration)
				assert.Equal(t, tt.expected.MetricsPort, config.MetricsPort)
				assert.Equal(t, tt.expected.DatabaseURL, config.DatabaseURL)
			}
		})
	}
}

func TestLoadConfig_EnvFileError(t *testing.T) {
	os.Remove(".env")

	config, err := LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "error loading .env file")
}
