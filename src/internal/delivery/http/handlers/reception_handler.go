package handlers

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PVZHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	var req request.CreateReceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.PVZID == "" {
		h.sendError(w, "ID ПВЗ обязателен", http.StatusBadRequest)
		return
	}

	pvzID, err := uuid.Parse(req.PVZID)
	if err != nil {
		h.sendError(w, "Неверный формат ID ПВЗ", http.StatusBadRequest)
		return
	}

	reception, err := h.pvzService.CreateReception(pvzID)
	if err != nil {
		switch err {
		case apperrors.ErrPVZNotFound:
			h.sendError(w, "ПВЗ не найден", http.StatusBadRequest)
		case apperrors.ErrActiveReceptionExists:
			h.sendError(w, "Уже есть активная приемка", http.StatusBadRequest)
		default:
			log.Println(err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reception)
}

func (h *PVZHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) {
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

	reception, err := h.pvzService.CloseLastReception(pvzID)
	if err != nil {
		switch err {
		case apperrors.ErrPVZNotFound:
			h.sendError(w, "ПВЗ не найден", http.StatusBadRequest)
		case apperrors.ErrNoActiveReception:
			h.sendError(w, "Нет активной приемки", http.StatusBadRequest)
		case apperrors.ErrReceptionAlreadyClosed:
			h.sendError(w, "Приемка уже закрыта", http.StatusBadRequest)
		default:
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reception)
}
