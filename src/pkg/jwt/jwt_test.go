package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestTokenManager_GenerateToken(t *testing.T) {
	manager := NewTokenManager("test-secret")

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{
			name:    "Valid Token Generation",
			role:    "employee",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateToken(tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)

				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret"), nil
				})

				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)

				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				assert.True(t, ok)
				assert.Equal(t, tt.role, claims["role"])

				exp := time.Unix(int64(claims["exp"].(float64)), 0)
				assert.True(t, exp.After(time.Now()))
			}
		})
	}
}
