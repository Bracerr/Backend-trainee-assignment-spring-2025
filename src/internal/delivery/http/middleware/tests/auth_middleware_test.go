package middleware_test

import (
	"avito-backend/src/internal/delivery/http/ctxkeys"
	"avito-backend/src/internal/delivery/http/middleware"
	"avito-backend/src/pkg/jwt"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	tokenManager := jwt.NewTokenManager("test-secret", "24h")

	tests := []struct {
		name           string
		setupAuth      func(r *http.Request)
		expectedStatus int
		expectedMsg    string
		checkRole      bool
		expectedRole   string
	}{
		{
			name: "Success with Valid Token",
			setupAuth: func(r *http.Request) {
				token, _ := tokenManager.GenerateToken("admin")
				r.Header.Set("Authorization", "Bearer "+token)
			},
			expectedStatus: http.StatusOK,
			checkRole:      true,
			expectedRole:   "admin",
		},
		{
			name:           "No Authorization Header",
			setupAuth:      func(r *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Отсутствует токен авторизации",
		},
		{
			name: "Invalid Authorization Format",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "InvalidFormat")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Неверный формат токена",
		},
		{
			name: "Bearer Without Token",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Неверный формат токена",
		},
		{
			name: "Invalid Token",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer invalid.token.here")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Неверный токен",
		},
		{
			name: "Token with Wrong Format",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer abc")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Неверный токен",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkRole {
					role := r.Context().Value(ctxkeys.UserRoleKey).(string)
					assert.Equal(t, tt.expectedRole, role)
				}
				w.WriteHeader(http.StatusOK)
			})

			middleware := middleware.AuthMiddleware(tokenManager)(nextHandler)

			req := httptest.NewRequest("GET", "/", nil)
			tt.setupAuth(req)

			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus != http.StatusOK {
				var response struct {
					Message string `json:"message"`
				}
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMsg, response.Message)
			}
		})
	}
}

func TestAuthMiddleware_ContextPropagation(t *testing.T) {
	tokenManager := jwt.NewTokenManager("test-secret", "24h")
	token, err := tokenManager.GenerateToken("admin")
	require.NoError(t, err)

	var capturedRole string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRole = r.Context().Value(ctxkeys.UserRoleKey).(string)
		w.WriteHeader(http.StatusOK)
	})

	middleware := middleware.AuthMiddleware(tokenManager)(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "admin", capturedRole)
}

func TestAuthMiddleware_ResponseHeaders(t *testing.T) {
	tokenManager := jwt.NewTokenManager("test-secret", "24h")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := middleware.AuthMiddleware(tokenManager)(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}
