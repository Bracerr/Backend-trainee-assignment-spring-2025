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

func (m *MockPVZService) DeleteLastProduct(pvzID uuid.UUID) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func (m *MockPVZService) CloseLastReception(pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockPVZService) GetPVZsWithReceptions(startDate, endDate time.Time, offset, limit int) ([]*models.PVZWithReceptions, error) {
	args := m.Called(startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PVZWithReceptions), args.Error(1)
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


func TestPVZHandler_GetPVZs(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  string
		mockBehavior func(s *MockPVZService)
		expectedCode int
		expectedBody string
	}{
		{
			name:        "Success",
			queryParams: "startDate=2023-01-01T00:00:00Z&endDate=2023-12-31T23:59:59Z&page=1&limit=10",
			mockBehavior: func(s *MockPVZService) {
				s.On("GetPVZsWithReceptions",
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					0,
					10,
				).Return(
					[]*models.PVZWithReceptions{
						{
							PVZ: &models.PVZ{
								ID:               uuid.New(),
								RegistrationDate: time.Now(),
								City:             models.City("Москва"),
							},
							Receptions: make([]models.ReceptionWithProducts, 0),
						},
					},
					nil,
				)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:        "Invalid Start Date Format",
			queryParams: "startDate=2023/01/01&endDate=2023-12-31T23:59:59Z&page=1&limit=10",
			mockBehavior: func(s *MockPVZService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Неверный формат начальной даты\"}\n",
		},
		{
			name:        "Invalid End Date Format",
			queryParams: "startDate=2023-01-01T00:00:00Z&endDate=2023/12/31&page=1&limit=10",
			mockBehavior: func(s *MockPVZService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Неверный формат конечной даты\"}\n",
		},
		{
			name:        "Invalid Page",
			queryParams: "startDate=2023-01-01T00:00:00Z&endDate=2023-12-31T23:59:59Z&page=0&limit=10",
			mockBehavior: func(s *MockPVZService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Неверный номер страницы\"}\n",
		},
		{
			name:        "Invalid Limit",
			queryParams: "startDate=2023-01-01T00:00:00Z&endDate=2023-12-31T23:59:59Z&page=1&limit=31",
			mockBehavior: func(s *MockPVZService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Неверное количество элементов на странице\"}\n",
		},
		{
			name:        "Invalid Date Range",
			queryParams: "startDate=2023-12-31T00:00:00Z&endDate=2023-01-01T00:00:00Z&page=1&limit=10",
			mockBehavior: func(s *MockPVZService) {
				s.On("GetPVZsWithReceptions", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0, 10).Return(
					nil, apperrors.ErrInvalidDateRange)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Неверный диапазон дат\"}\n",
		},
		{
			name:        "Service Error",
			queryParams: "startDate=2023-01-01T00:00:00Z&endDate=2023-12-31T23:59:59Z&page=1&limit=10",
			mockBehavior: func(s *MockPVZService) {
				s.On("GetPVZsWithReceptions", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0, 10).Return(
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

			req := httptest.NewRequest("GET", "/pvz?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.GetPVZs(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}
