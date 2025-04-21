package middleware_test

import (
	"avito-backend/src/internal/delivery/http/ctxkeys"
	"avito-backend/src/internal/delivery/http/middleware"
	"avito-backend/src/internal/domain/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		requiredRole   models.Role
		contextRole    string
		expectedStatus int
	}{
		{
			name:           "Success - Matching Role",
			requiredRole:   models.EmployeeRole,
			contextRole:    string(models.EmployeeRole),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure - Wrong Role",
			requiredRole:   models.ModeratorRole,
			contextRole:    string(models.EmployeeRole),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Failure - No Role",
			requiredRole:   models.ModeratorRole,
			contextRole:    "",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := middleware.RequireRole(tt.requiredRole)(nextHandler)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.contextRole != "" {
				ctx := context.WithValue(req.Context(), ctxkeys.UserRoleKey, tt.contextRole)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusForbidden {
				var response struct {
					Message string `json:"message"`
				}
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "Доступ запрещен", response.Message)
			}
		})
	}
}

func TestRequireRoles(t *testing.T) {
	tests := []struct {
		name           string
		requiredRoles  []models.Role
		contextRole    string
		expectedStatus int
	}{
		{
			name:           "Success - Role in List",
			requiredRoles:  []models.Role{models.ModeratorRole, models.EmployeeRole},
			contextRole:    string(models.ModeratorRole),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success - Another Role in List",
			requiredRoles:  []models.Role{models.ModeratorRole, models.EmployeeRole},
			contextRole:    string(models.EmployeeRole),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure - Role Not in List",
			requiredRoles:  []models.Role{models.ModeratorRole, models.EmployeeRole},
			contextRole:    "unknown_role",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Failure - No Role",
			requiredRoles:  []models.Role{models.ModeratorRole, models.EmployeeRole},
			contextRole:    "",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := middleware.RequireRoles(tt.requiredRoles)(nextHandler)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.contextRole != "" {
				ctx := context.WithValue(req.Context(), ctxkeys.UserRoleKey, tt.contextRole)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusForbidden {
				var response struct {
					Message string `json:"message"`
				}
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "Доступ запрещен", response.Message)
			}
		})
	}
}