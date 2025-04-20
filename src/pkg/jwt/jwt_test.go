package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenManager(t *testing.T) {
	manager := NewTokenManager("test-key", "1h")
	assert.NotNil(t, manager)
	assert.Equal(t, "test-key", manager.signingKey)
	assert.Equal(t, "1h", manager.duration)
}

func TestTokenManager_GenerateToken(t *testing.T) {
	tests := []struct {
		name       string
		signingKey string
		duration   string
		role       string
		wantErr    bool
	}{
		{
			name:       "Success generate token",
			signingKey: "test-key",
			duration:   "1h",
			role:       "admin",
			wantErr:    false,
		},
		{
			name:       "Invalid duration",
			signingKey: "test-key",
			duration:   "invalid",
			role:       "admin",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewTokenManager(tt.signingKey, tt.duration)
			token, err := manager.GenerateToken(tt.role)

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

func TestTokenManager_ValidateToken(t *testing.T) {
	signingKey := "test-key"
	duration := "1h"
	manager := NewTokenManager(signingKey, duration)

	tests := []struct {
		name        string
		setupToken  func() string
		wantRole    string
		wantErr     bool
		errorString string
	}{
		{
			name: "Valid token",
			setupToken: func() string {
				token, _ := manager.GenerateToken("admin")
				return token
			},
			wantRole: "admin",
			wantErr:  false,
		},
		{
			name: "Expired token",
			setupToken: func() string {
				expiredManager := NewTokenManager(signingKey, "-1h")
				token, _ := expiredManager.GenerateToken("admin")
				return token
			},
			wantRole:    "",
			wantErr:     true,
			errorString: "token is expired",
		},
		{
			name: "Invalid token format",
			setupToken: func() string {
				return "invalid.token.format"
			},
			wantRole: "",
			wantErr:  true,
		},
		{
			name: "Empty token",
			setupToken: func() string {
				return ""
			},
			wantRole: "",
			wantErr:  true,
		},
		{
			name: "Token with other signing key",
			setupToken: func() string {
				otherManager := NewTokenManager("other-key", duration)
				token, _ := otherManager.GenerateToken("admin")
				return token
			},
			wantRole: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			role, err := manager.ValidateToken(token)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorString != "" {
					assert.Contains(t, err.Error(), tt.errorString)
				}
				assert.Empty(t, role)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRole, role)
			}
		})
	}
}

func TestTokenManager_TokenExpiration(t *testing.T) {
	manager := NewTokenManager("test-key", "1s")
	token, err := manager.GenerateToken("admin")
	assert.NoError(t, err)

	role, err := manager.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "admin", role)

	time.Sleep(2 * time.Second)

	role, err = manager.ValidateToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
	assert.Empty(t, role)
}