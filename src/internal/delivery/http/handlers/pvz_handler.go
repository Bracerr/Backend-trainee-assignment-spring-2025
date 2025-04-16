package handlers

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/service"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type PVZHandler struct {
	pvzService service.PVZServiceInterface
}

func NewPVZHandler(pvzService service.PVZServiceInterface) *PVZHandler {
	return &PVZHandler{
		pvzService: pvzService,
	}
}

func (h *PVZHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreatePVZRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.City == "" {
		h.sendError(w, "Город не может быть пустым", http.StatusBadRequest)
		return
	}

	pvz, err := h.pvzService.Create(req.City)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidCity:
			h.sendError(w, "Недопустимый город", http.StatusBadRequest)
		default:
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pvz)
}

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
		case apperrors.ErrActiveReceptionExists:
			h.sendError(w, "Уже есть активная приемка", http.StatusBadRequest)
		default:
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reception)
}

func (h *PVZHandler) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
