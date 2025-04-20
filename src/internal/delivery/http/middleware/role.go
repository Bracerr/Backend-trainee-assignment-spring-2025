package middleware

import (
	"avito-backend/src/internal/delivery/http/ctxkeys"
	"avito-backend/src/internal/delivery/http/dto/response"
	"avito-backend/src/internal/domain/models"
	"encoding/json"
	"net/http"
)

func RequireRole(role models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value(ctxkeys.UserRoleKey)

			if userRole == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Доступ запрещен"})
				return
			}

			roleStr := userRole.(string)

			if roleStr != string(role) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Доступ запрещен"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRoles(roles []models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value(ctxkeys.UserRoleKey)

			if userRole == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Доступ запрещен"})
				return
			}

			roleStr := userRole.(string)
			hasRole := false
			for _, role := range roles {
				if roleStr == string(role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Доступ запрещен"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
