package handlers_test

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/ctxkeys"
	"avito-backend/src/internal/delivery/http/handlers"
	"avito-backend/src/internal/domain/models"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPVZHandler_CreateReception(t *testing.T) {
	tests := []struct {
		name         string
		pvzID        string
		mockBehavior func(s *MockPVZService)
		expectedCode int
		expectedBody string
	}{
		{
			name:  "Success",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("CreateReception", mock.AnythingOfType("uuid.UUID")).Return(&models.Reception{
					ID:       uuid.New(),
					DateTime: time.Now(),
					PVZID:    uuid.New(),
					Status:   models.InProgress,
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Empty PVZ ID",
			pvzID:        "",
			mockBehavior: func(s *MockPVZService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"ID ПВЗ обязателен\"}\n",
		},
		{
			name:         "Invalid PVZ ID Format",
			pvzID:        "invalid-uuid",
			mockBehavior: func(s *MockPVZService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Неверный формат ID ПВЗ\"}\n",
		},
		{
			name:  "Active Reception Exists",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("CreateReception", mock.AnythingOfType("uuid.UUID")).Return(nil, apperrors.ErrActiveReceptionExists)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Уже есть активная приемка\"}\n",
		},
		{
			name:  "Service Error",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("CreateReception", mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"message\":\"Внутренняя ошибка сервера\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.mockBehavior(mockService)
			handler := handlers.NewPVZHandler(mockService)

			body := []byte(fmt.Sprintf(`{"pvzId":"%s"}`, tt.pvzID))
			req := httptest.NewRequest("POST", "/pvz/receptions", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), ctxkeys.UserRoleKey, string(models.EmployeeRole))
			req = req.WithContext(ctx)

			handler.CreateReception(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_CreateReception_InvalidJSON(t *testing.T) {
	mockService := new(MockPVZService)
	handler := handlers.NewPVZHandler(mockService)

	req := httptest.NewRequest("POST", "/pvz/receptions", bytes.NewBuffer([]byte(`{"pvzId": invalid_json`)))
	w := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), ctxkeys.UserRoleKey, string(models.EmployeeRole))
	req = req.WithContext(ctx)

	handler.CreateReception(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"message\":\"Неверный формат запроса\"}\n", w.Body.String())
}


func TestPVZHandler_CloseLastReception(t *testing.T) {
	tests := []struct {
		name         string
		pvzID        string
		mockBehavior func(s *MockPVZService)
		expectedCode int
		expectedBody string
	}{
		{
			name:  "Success",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("CloseLastReception", mock.AnythingOfType("uuid.UUID")).Return(
					&models.Reception{
						ID:       uuid.New(),
						DateTime: time.Now(),
						Status:   models.Closed,
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Empty PVZ ID",
			pvzID:        "",
			mockBehavior: func(s *MockPVZService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"ID ПВЗ обязателен\"}\n",
		},
		{
			name:         "Invalid PVZ ID Format",
			pvzID:        "invalid-uuid",
			mockBehavior: func(s *MockPVZService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Неверный формат ID ПВЗ\"}\n",
		},
		{
			name:  "PVZ Not Found",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("CloseLastReception", mock.AnythingOfType("uuid.UUID")).Return(
					nil, apperrors.ErrPVZNotFound)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"ПВЗ не найден\"}\n",
		},
		{
			name:  "No Active Reception",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("CloseLastReception", mock.AnythingOfType("uuid.UUID")).Return(
					nil, apperrors.ErrNoActiveReception)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Нет активной приемки\"}\n",
		},
		{
			name:  "Reception Already Closed",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("CloseLastReception", mock.AnythingOfType("uuid.UUID")).Return(
					nil, apperrors.ErrReceptionAlreadyClosed)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Приемка уже закрыта\"}\n",
		},
		{
			name:  "Service Error",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("CloseLastReception", mock.AnythingOfType("uuid.UUID")).Return(
					nil, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"message\":\"Внутренняя ошибка сервера\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.mockBehavior(mockService)
			handler := handlers.NewPVZHandler(mockService)

			req := httptest.NewRequest("POST", fmt.Sprintf("/pvz/%s/close_last_reception", tt.pvzID), nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("pvzId", tt.pvzID)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, ctxkeys.UserRoleKey, string(models.EmployeeRole))
			req = req.WithContext(ctx)

			handler.CloseLastReception(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}
