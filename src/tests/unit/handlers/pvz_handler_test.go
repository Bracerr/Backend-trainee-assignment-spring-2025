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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPVZService struct {
	mock.Mock
}

func (m *MockPVZService) Create(city string) (*models.PVZ, error) {
	args := m.Called(city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *MockPVZService) CreateReception(pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockPVZService) CreateProduct(pvzID uuid.UUID, productType string) (*models.Product, error) {
	args := m.Called(pvzID, productType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func TestPVZHandler_Create(t *testing.T) {
	tests := []struct {
		name         string
		input        request.CreatePVZRequest
		mockBehavior func(s *MockPVZService)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success",
			input: request.CreatePVZRequest{
				City: "Москва",
			},
			mockBehavior: func(s *MockPVZService) {
				s.On("Create", "Москва").Return(&models.PVZ{
					ID:               uuid.New(),
					RegistrationDate: time.Now(),
					City:             "Москва",
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Invalid City",
			input: request.CreatePVZRequest{
				City: "Новосибирск",
			},
			mockBehavior: func(s *MockPVZService) {
				s.On("Create", "Новосибирск").Return(nil, apperrors.ErrInvalidCity)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Недопустимый город\"}\n",
		},
		{
			name: "Service Error",
			input: request.CreatePVZRequest{
				City: "Москва",
			},
			mockBehavior: func(s *MockPVZService) {
				s.On("Create", "Москва").Return(nil, errors.New("service error"))
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

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), ctxkeys.UserRoleKey, string(models.ModeratorRole))
			req = req.WithContext(ctx)

			handler.Create(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_Create_InvalidJSON(t *testing.T) {
	mockService := new(MockPVZService)
	handler := handlers.NewPVZHandler(mockService)

	req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer([]byte(`{"city": invalid_json`)))
	w := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), ctxkeys.UserRoleKey, string(models.ModeratorRole))
	req = req.WithContext(ctx)

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"message\":\"Неверный формат запроса\"}\n", w.Body.String())
}

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
