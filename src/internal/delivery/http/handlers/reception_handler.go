package handlers

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/pkg/metrics"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"avito-backend/src/pkg/logger"
	"log/slog"
)

func (h *PVZHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.CreateReceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "ошибка декодирования запроса", "error", err)
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.PVZID == "" {
		slog.WarnContext(ctx, "ID ПВЗ не указан")
		h.sendError(w, "ID ПВЗ обязателен", http.StatusBadRequest)
		return
	}

	pvzID, err := uuid.Parse(req.PVZID)
	if err != nil {
		slog.WarnContext(ctx, "неверный формат ID ПВЗ")
		h.sendError(w, "Неверный формат ID ПВЗ", http.StatusBadRequest)
		return
	}

	ctx = logger.WithPVZID(ctx, pvzID.String())
	slog.InfoContext(ctx, "создание приемки")

	reception, err := h.pvzService.CreateReception(pvzID)
	if err != nil {
		switch err {
		case apperrors.ErrPVZNotFound:
			slog.WarnContext(ctx, "ПВЗ не найден")
			h.sendError(w, "ПВЗ не найден", http.StatusBadRequest)
		case apperrors.ErrActiveReceptionExists:
			slog.WarnContext(ctx, "уже есть активная приемка")
			h.sendError(w, "Уже есть активная приемка", http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "ошибка создания приемки", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	metrics.ReceptionsCreatedTotal.Inc()
	slog.InfoContext(ctx, "приемка создана", "reception_id", reception.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reception)
}

func (h *PVZHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pvzIDStr := chi.URLParam(r, "pvzId")
	if pvzIDStr == "" {
		slog.WarnContext(ctx, "не указан ID ПВЗ")
		h.sendError(w, "ID ПВЗ обязателен", http.StatusBadRequest)
		return
	}

	pvzID, err := uuid.Parse(pvzIDStr)
	if err != nil {
		slog.WarnContext(ctx, "неверный формат ID ПВЗ")
		h.sendError(w, "Неверный формат ID ПВЗ", http.StatusBadRequest)
		return
	}

	ctx = logger.WithPVZID(ctx, pvzID.String())
	slog.InfoContext(ctx, "закрытие последней приемки")

	reception, err := h.pvzService.CloseLastReception(pvzID)
	if err != nil {
		switch err {
		case apperrors.ErrPVZNotFound:
			slog.WarnContext(ctx, "ПВЗ не найден")
			h.sendError(w, "ПВЗ не найден", http.StatusBadRequest)
		case apperrors.ErrNoActiveReception:
			slog.WarnContext(ctx, "нет активной приемки")
			h.sendError(w, "Нет активной приемки", http.StatusBadRequest)
		case apperrors.ErrReceptionAlreadyClosed:
			slog.WarnContext(ctx, "приемка уже закрыта")
			h.sendError(w, "Приемка уже закрыта", http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "ошибка закрытия приемки", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	slog.InfoContext(ctx, "приемка закрыта", "reception_id", reception.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reception)
}
