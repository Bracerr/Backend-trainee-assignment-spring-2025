package middleware

import (
	"avito-backend/src/internal/delivery/http/ctxkeys"
	"avito-backend/src/internal/delivery/http/dto/response"
	"avito-backend/src/internal/domain/models"
	"encoding/json"
	"net/http"
)

func RequireRole(role models.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(ctxkeys.UserRoleKey).(string)
			if !ok || userRole == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response.ErrorResponse{
					Message: "Отсутствует роль пользователя",
				})
				return
			}

			if userRole != string(role) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(response.ErrorResponse{
					Message: "Доступ запрещен",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
