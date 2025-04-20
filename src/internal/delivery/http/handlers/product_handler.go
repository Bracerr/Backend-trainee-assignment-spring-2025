package handlers

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/pkg/logger"
	"avito-backend/src/pkg/metrics"
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PVZHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.CreateProductRequest
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

	if req.Type == "" {
		slog.WarnContext(ctx, "тип товара не указан")
		h.sendError(w, "Тип товара обязателен", http.StatusBadRequest)
		return
	}

	pvzID, err := uuid.Parse(req.PVZID)
	if err != nil {
		slog.WarnContext(ctx, "неверный формат ID ПВЗ")
		h.sendError(w, "Неверный формат ID ПВЗ", http.StatusBadRequest)
		return
	}

	ctx = logger.WithPVZID(ctx, pvzID.String())

	slog.InfoContext(ctx, "создание товара", "type", req.Type)

	product, err := h.pvzService.CreateProduct(pvzID, req.Type)
	if err != nil {
		switch err {
		case apperrors.ErrNoActiveReception:
			slog.WarnContext(ctx, "нет активной приемки")
			h.sendError(w, "Нет активной приемки", http.StatusBadRequest)
		case apperrors.ErrPVZNotFound:
			slog.WarnContext(ctx, "ПВЗ не найден")
			h.sendError(w, "ПВЗ не найден", http.StatusBadRequest)
		case apperrors.ErrInvalidProductType:
			slog.WarnContext(ctx, "недопустимый тип товара")
			h.sendError(w, "Недопустимый тип товара", http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "ошибка создания товара", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	metrics.ProductsAddedTotal.Inc()
	slog.InfoContext(ctx, "товар создан", "product_id", product.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *PVZHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
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

	slog.InfoContext(ctx, "удаление последнего товара")

	err = h.pvzService.DeleteLastProduct(pvzID)
	if err != nil {
		switch err {
		case apperrors.ErrPVZNotFound:
			slog.WarnContext(ctx, "ПВЗ не найден")
			h.sendError(w, "ПВЗ не найден", http.StatusBadRequest)
		case apperrors.ErrNoActiveReception:
			slog.WarnContext(ctx, "нет активной приемки")
			h.sendError(w, "Нет активной приемки", http.StatusBadRequest)
		case apperrors.ErrReceptionClosed:
			slog.WarnContext(ctx, "приемка закрыта")
			h.sendError(w, "Приемка уже закрыта", http.StatusBadRequest)
		case apperrors.ErrNoProductsToDelete:
			slog.WarnContext(ctx, "нет товаров для удаления")
			h.sendError(w, "Нет товаров для удаления", http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "ошибка удаления товара", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	slog.InfoContext(ctx, "товар удален")
	w.WriteHeader(http.StatusOK)
}
