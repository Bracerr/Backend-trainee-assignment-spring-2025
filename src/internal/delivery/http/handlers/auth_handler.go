package handlers

import (
	"encoding/json"
	"net/http"

	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/delivery/http/dto/response"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/service"
)

type AuthHandler struct {
	authService service.AuthServiceInterface
}

func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req request.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.Role != string(models.EmployeeRole) && req.Role != string(models.ModeratorRole) {
		h.sendError(w, "Недопустимая роль", http.StatusBadRequest)
		return
	}

	token, err := h.authService.GenerateToken(req.Role)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (h *AuthHandler) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response.ErrorResponse{
		Message: message,
	})
}
