package handlers

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/service"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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

func (h *PVZHandler) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
