package handlers

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/pkg/metrics"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PVZHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req request.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.PVZID == "" {
		h.sendError(w, "ID ПВЗ обязателен", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		h.sendError(w, "Тип товара обязателен", http.StatusBadRequest)
		return
	}

	pvzID, err := uuid.Parse(req.PVZID)
	if err != nil {
		h.sendError(w, "Неверный формат ID ПВЗ", http.StatusBadRequest)
		return
	}

	product, err := h.pvzService.CreateProduct(pvzID, req.Type)
	if err != nil {
		switch err {
		case apperrors.ErrNoActiveReception:
			h.sendError(w, "Нет активной приемки", http.StatusBadRequest)
		case apperrors.ErrPVZNotFound:
			h.sendError(w, "ПВЗ не найден", http.StatusBadRequest)
		case apperrors.ErrInvalidProductType:
			h.sendError(w, "Недопустимый тип товара", http.StatusBadRequest)
		default:
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	metrics.ProductsAddedTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *PVZHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	pvzIDStr := chi.URLParam(r, "pvzId")
	if pvzIDStr == "" {
		h.sendError(w, "ID ПВЗ обязателен", http.StatusBadRequest)
		return
	}

	pvzID, err := uuid.Parse(pvzIDStr)
	if err != nil {
		h.sendError(w, "Неверный формат ID ПВЗ", http.StatusBadRequest)
		return
	}

	err = h.pvzService.DeleteLastProduct(pvzID)
	if err != nil {
		switch err {
		case apperrors.ErrPVZNotFound:
			h.sendError(w, "ПВЗ не найден", http.StatusBadRequest)
		case apperrors.ErrNoActiveReception:
			h.sendError(w, "Нет активной приемки", http.StatusBadRequest)
		case apperrors.ErrReceptionClosed:
			h.sendError(w, "Приемка уже закрыта", http.StatusBadRequest)
		case apperrors.ErrNoProductsToDelete:
			h.sendError(w, "Нет товаров для удаления", http.StatusBadRequest)
		default:
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
