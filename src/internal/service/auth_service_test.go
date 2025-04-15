package service

import (
	"testing"

	"avito-backend/src/pkg/jwt"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_GenerateToken(t *testing.T) {
	tokenManager := jwt.NewTokenManager("test-secret")
	service := NewAuthService(tokenManager)

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{
			name:    "Success Employee",
			role:    "employee",
			wantErr: false,
		},
		{
			name:    "Success Moderator",
			role:    "moderator",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}
