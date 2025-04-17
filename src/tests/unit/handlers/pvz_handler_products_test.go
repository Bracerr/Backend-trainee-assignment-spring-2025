package handlers_test

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/ctxkeys"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/delivery/http/handlers"
	"avito-backend/src/internal/domain/models"
	"bytes"
	"context"
	"encoding/json"
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

func TestPVZHandler_CreateProduct(t *testing.T) {
	tests := []struct {
		name         string
		input        request.CreateProductRequest
		mockBehavior func(s *MockPVZService)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success",
			input: request.CreateProductRequest{
				Type:  string(models.Electronics),
				PVZID: uuid.New().String(),
			},
			mockBehavior: func(s *MockPVZService) {
				s.On("CreateProduct",
					mock.AnythingOfType("uuid.UUID"),
					string(models.Electronics),
				).Return(&models.Product{
					ID:          uuid.New(),
					DateTime:    time.Now(),
					Type:        models.Electronics,
					ReceptionID: uuid.New(),
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Empty PVZ ID",
			input: request.CreateProductRequest{
				Type:  string(models.Electronics),
				PVZID: "",
			},
			mockBehavior: func(s *MockPVZService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"ID ПВЗ обязателен\"}\n",
		},
		{
			name: "Empty Product Type",
			input: request.CreateProductRequest{
				Type:  "",
				PVZID: uuid.New().String(),
			},
			mockBehavior: func(s *MockPVZService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Тип товара обязателен\"}\n",
		},
		{
			name: "Invalid PVZ ID Format",
			input: request.CreateProductRequest{
				Type:  string(models.Electronics),
				PVZID: "invalid-uuid",
			},
			mockBehavior: func(s *MockPVZService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Неверный формат ID ПВЗ\"}\n",
		},
		{
			name: "PVZ Not Found",
			input: request.CreateProductRequest{
				Type:  string(models.Electronics),
				PVZID: uuid.New().String(),
			},
			mockBehavior: func(s *MockPVZService) {
				s.On("CreateProduct",
					mock.AnythingOfType("uuid.UUID"),
					string(models.Electronics),
				).Return(nil, apperrors.ErrPVZNotFound)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"ПВЗ не найден\"}\n",
		},
		{
			name: "No Active Reception",
			input: request.CreateProductRequest{
				Type:  string(models.Electronics),
				PVZID: uuid.New().String(),
			},
			mockBehavior: func(s *MockPVZService) {
				s.On("CreateProduct",
					mock.AnythingOfType("uuid.UUID"),
					string(models.Electronics),
				).Return(nil, apperrors.ErrNoActiveReception)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Нет активной приемки\"}\n",
		},
		{
			name: "Invalid Product Type",
			input: request.CreateProductRequest{
				Type:  "invalid_type",
				PVZID: uuid.New().String(),
			},
			mockBehavior: func(s *MockPVZService) {
				s.On("CreateProduct",
					mock.AnythingOfType("uuid.UUID"),
					"invalid_type",
				).Return(nil, apperrors.ErrInvalidProductType)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Недопустимый тип товара\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.mockBehavior(mockService)
			handler := handlers.NewPVZHandler(mockService)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), ctxkeys.UserRoleKey, string(models.EmployeeRole))
			req = req.WithContext(ctx)

			handler.CreateProduct(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_CreateProduct_InvalidJSON(t *testing.T) {
	mockService := new(MockPVZService)
	handler := handlers.NewPVZHandler(mockService)

	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer([]byte(`{"type": invalid_json`)))
	w := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), ctxkeys.UserRoleKey, string(models.EmployeeRole))
	req = req.WithContext(ctx)

	handler.CreateProduct(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"message\":\"Неверный формат запроса\"}\n", w.Body.String())
}

func TestPVZHandler_DeleteLastProduct(t *testing.T) {
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
				s.On("DeleteLastProduct", mock.AnythingOfType("uuid.UUID")).Return(nil)
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
				s.On("DeleteLastProduct", mock.AnythingOfType("uuid.UUID")).Return(apperrors.ErrPVZNotFound)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"ПВЗ не найден\"}\n",
		},
		{
			name:  "No Active Reception",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("DeleteLastProduct", mock.AnythingOfType("uuid.UUID")).Return(apperrors.ErrNoActiveReception)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Нет активной приемки\"}\n",
		},
		{
			name:  "Reception Closed",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("DeleteLastProduct", mock.AnythingOfType("uuid.UUID")).Return(apperrors.ErrReceptionClosed)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Приемка уже закрыта\"}\n",
		},
		{
			name:  "No Products in Reception",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("DeleteLastProduct", mock.AnythingOfType("uuid.UUID")).Return(apperrors.ErrNoProductsToDelete)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Нет товаров для удаления\"}\n",
		},
		{
			name:  "Service Error",
			pvzID: uuid.New().String(),
			mockBehavior: func(s *MockPVZService) {
				s.On("DeleteLastProduct", mock.AnythingOfType("uuid.UUID")).Return(errors.New("service error"))
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

			req := httptest.NewRequest("POST", fmt.Sprintf("/pvz/%s/delete_last_product", tt.pvzID), nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("pvzId", tt.pvzID)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, ctxkeys.UserRoleKey, string(models.EmployeeRole))
			req = req.WithContext(ctx)

			handler.DeleteLastProduct(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}
