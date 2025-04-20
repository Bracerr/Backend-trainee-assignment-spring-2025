package middleware

import (
	"avito-backend/src/internal/delivery/http/ctxkeys"
	"avito-backend/src/internal/delivery/http/dto/response"
	"avito-backend/src/pkg/jwt"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func AuthMiddleware(tokenManager *jwt.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response.ErrorResponse{
					Message: "Отсутствует токен авторизации",
				})
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response.ErrorResponse{
					Message: "Неверный формат токена",
				})
				return
			}

			userRole, err := tokenManager.ValidateToken(headerParts[1])
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response.ErrorResponse{
					Message: "Неверный токен",
				})
				return
			}

			ctx := context.WithValue(r.Context(), ctxkeys.UserRoleKey, userRole)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
