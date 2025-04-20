package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/mail"

	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/delivery/http/dto/response"
	"avito-backend/src/internal/service"
	"avito-backend/src/pkg/logger"
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
	ctx := r.Context()

	var req request.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "ошибка декодирования запроса", "error", err)
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "попытка dummy login", "role", req.Role)

	token, err := h.authService.GenerateToken(req.Role)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidRole:
			slog.WarnContext(ctx, "указана недопустимая роль", "role", req.Role)
			h.sendError(w, "Недопустимая роль", http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "ошибка генерации токена", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	slog.InfoContext(ctx, "токен успешно создан")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "ошибка декодирования запроса", "error", err)
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "попытка регистрации",
		"email", req.Email,
		"role", req.Role)

	if req.Email == "" || req.Role == "" || req.Password == "" {
		slog.WarnContext(ctx, "отсутствуют обязательные поля")
		h.sendError(w, "Отсутствуют обязательные поля", http.StatusBadRequest)
		return
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		slog.WarnContext(ctx, "неверный формат email", "email", req.Email)
		h.sendError(w, "Неверный формат email", http.StatusBadRequest)
		return
	}

	ctx = logger.WithEmail(ctx, req.Email)
	user, err := h.authService.Register(req.Email, req.Password, req.Role)
	if err != nil {
		switch err {
		case apperrors.ErrUserAlreadyExists:
			slog.WarnContext(ctx, "пользователь уже существует")
			h.sendError(w, "Пользователь уже существует", http.StatusBadRequest)
		case apperrors.ErrInvalidRole:
			slog.WarnContext(ctx, "недопустимая роль", "role", req.Role)
			h.sendError(w, "Недопустимая роль", http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "ошибка регистрации", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	slog.InfoContext(ctx, "регистрация успешна")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "ошибка декодирования запроса", "error", err)
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "попытка входа", "email", req.Email)

	if req.Email == "" || req.Password == "" {
		slog.WarnContext(ctx, "отсутствуют учетные данные")
		h.sendError(w, "Отсутствуют учетные данные", http.StatusUnauthorized)
		return
	}

	ctx = logger.WithEmail(ctx, req.Email)
	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidCredentials:
			slog.WarnContext(ctx, "неверные учетные данные")
			h.sendError(w, "Неверные учетные данные", http.StatusUnauthorized)
		default:
			slog.ErrorContext(ctx, "ошибка авторизации", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	slog.InfoContext(ctx, "вход успешен")
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
