package handlers

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/service"
	"avito-backend/src/pkg/metrics"
	"encoding/json"
	"log/slog"
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
	ctx := r.Context()

	var req request.CreatePVZRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(ctx, "ошибка декодирования запроса", "error", err)
		h.sendError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if req.City == "" {
		slog.WarnContext(ctx, "город не указан")
		h.sendError(w, "Город не может быть пустым", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "создание ПВЗ", "city", req.City)

	pvz, err := h.pvzService.Create(req.City)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidCity:
			slog.WarnContext(ctx, "недопустимый город", "city", req.City)
			h.sendError(w, "Недопустимый город", http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "ошибка создания ПВЗ", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	metrics.PvzCreatedTotal.Inc()
	slog.InfoContext(ctx, "ПВЗ создан", "pvz_id", pvz.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pvz)
}

func (h *PVZHandler) GetPVZs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var startDate, endDate time.Time
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	if startDateStr != "" {
		var err error
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			slog.WarnContext(ctx, "неверный формат начальной даты", "start_date", startDateStr)
			h.sendError(w, "Неверный формат начальной даты", http.StatusBadRequest)
			return
		}
	}

	if endDateStr != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			slog.WarnContext(ctx, "неверный формат конечной даты", "end_date", endDateStr)
			h.sendError(w, "Неверный формат конечной даты", http.StatusBadRequest)
			return
		}
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			slog.WarnContext(ctx, "неверный номер страницы", "page", pageStr)
			h.sendError(w, "Неверный номер страницы", http.StatusBadRequest)
			return
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 30 {
			slog.WarnContext(ctx, "неверное количество элементов на странице", "limit", limitStr)
			h.sendError(w, "Неверное количество элементов на странице", http.StatusBadRequest)
			return
		}
	}

	offset := (page - 1) * limit

	slog.InfoContext(ctx, "получение списка ПВЗ",
		"start_date", startDateStr,
		"end_date", endDateStr,
		"page", page,
		"limit", limit)

	pvzs, err := h.pvzService.GetPVZsWithReceptions(startDate, endDate, offset, limit)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidDateRange:
			slog.WarnContext(ctx, "неверный диапазон дат")
			h.sendError(w, "Неверный диапазон дат", http.StatusBadRequest)
		case apperrors.ErrInvalidPagination:
			slog.WarnContext(ctx, "неверные параметры пагинации")
			h.sendError(w, "Неверные параметры пагинации", http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "ошибка получения списка ПВЗ", "error", err)
			h.sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	slog.InfoContext(ctx, "список ПВЗ получен", "count", len(pvzs))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pvzs)
}

func (h *PVZHandler) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
