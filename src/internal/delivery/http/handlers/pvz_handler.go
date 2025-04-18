package handlers

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/service"
	"avito-backend/src/pkg/metrics"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
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
	metrics.PvzCreatedTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pvz)
}

func (h *PVZHandler) GetPVZs(w http.ResponseWriter, r *http.Request) {
	var startDate, endDate time.Time

	if startDateStr := r.URL.Query().Get("startDate"); startDateStr != "" {
		var err error
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			h.sendError(w, "Неверный формат начальной даты", http.StatusBadRequest)
			return
		}
	}

	if endDateStr := r.URL.Query().Get("endDate"); endDateStr != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			h.sendError(w, "Неверный формат конечной даты", http.StatusBadRequest)
			return
		}
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			h.sendError(w, "Неверный номер страницы", http.StatusBadRequest)
			return
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 30 {
			h.sendError(w, "Неверное количество элементов на странице", http.StatusBadRequest)
			return
		}
	}

	offset := (page - 1) * limit

	pvzs, err := h.pvzService.GetPVZsWithReceptions(startDate, endDate, offset, limit)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidDateRange:
			h.sendError(w, "Неверный диапазон дат", http.StatusBadRequest)
		case apperrors.ErrInvalidPagination:
			h.sendError(w, "Неверные параметры пагинации", http.StatusBadRequest)
		default:
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pvzs)
}

func (h *PVZHandler) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
