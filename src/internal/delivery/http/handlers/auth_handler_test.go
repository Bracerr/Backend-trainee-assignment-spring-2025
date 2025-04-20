package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/delivery/http/handlers"
	"avito-backend/src/internal/domain/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(email, password, role string) (*models.User, error) {
	args := m.Called(email, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(email, password string) (string, error) {
	args := m.Called(email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateToken(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name         string
		input        request.RegisterRequest
		mockBehavior func(s *MockAuthService)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success",
			input: request.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "employee",
			},
			mockBehavior: func(s *MockAuthService) {
				s.On("Register", "test@example.com", "password123", "employee").Return(&models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
					Role:  "employee",
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "User Already Exists",
			input: request.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Role:     "employee",
			},
			mockBehavior: func(s *MockAuthService) {
				s.On("Register", "existing@example.com", "password123", "employee").Return(nil, apperrors.ErrUserAlreadyExists)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			input: request.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "employee",
			},
			mockBehavior: func(s *MockAuthService) {
				s.On("Register", "test@example.com", "password123", "employee").Return(nil, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"message\":\"Внутренняя ошибка сервера\"}\n",
		},
		{
			name: "Empty Email",
			input: request.RegisterRequest{
				Email:    "",
				Password: "password123",
				Role:     "employee",
			},
			mockBehavior: func(s *MockAuthService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Отсутствуют обязательные поля\"}\n",
		},
		{
			name: "Empty Password",
			input: request.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
				Role:     "employee",
			},
			mockBehavior: func(s *MockAuthService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Отсутствуют обязательные поля\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockBehavior(mockService)
			handler := handlers.NewAuthHandler(mockService)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name         string
		input        request.LoginRequest
		mockBehavior func(s *MockAuthService)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success",
			input: request.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockBehavior: func(s *MockAuthService) {
				s.On("Login", "test@example.com", "password123").Return("test-token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "\"test-token\"\n",
		},
		{
			name: "Invalid Credentials",
			input: request.LoginRequest{
				Email:    "wrong@example.com",
				Password: "wrongpass",
			},
			mockBehavior: func(s *MockAuthService) {
				s.On("Login", "wrong@example.com", "wrongpass").Return("", apperrors.ErrInvalidCredentials)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Service Error",
			input: request.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockBehavior: func(s *MockAuthService) {
				s.On("Login", "test@example.com", "password123").Return("", errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"message\":\"Внутренняя ошибка сервера\"}\n",
		},
		{
			name: "Empty Email",
			input: request.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			mockBehavior: func(s *MockAuthService) {
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "{\"message\":\"Отсутствуют учетные данные\"}\n",
		},
		{
			name: "Empty Password",
			input: request.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			mockBehavior: func(s *MockAuthService) {
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "{\"message\":\"Отсутствуют учетные данные\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockBehavior(mockService)
			handler := handlers.NewAuthHandler(mockService)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_DummyLogin(t *testing.T) {
	tests := []struct {
		name         string
		input        request.DummyLoginRequest
		mockBehavior func(s *MockAuthService, role string)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success Employee",
			input: request.DummyLoginRequest{
				Role: "employee",
			},
			mockBehavior: func(s *MockAuthService, role string) {
				s.On("GenerateToken", role).Return("test-token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "\"test-token\"\n",
		},
		{
			name: "Success Moderator",
			input: request.DummyLoginRequest{
				Role: "moderator",
			},
			mockBehavior: func(s *MockAuthService, role string) {
				s.On("GenerateToken", role).Return("test-token", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "\"test-token\"\n",
		},
		{
			name: "Invalid Role",
			input: request.DummyLoginRequest{
				Role: "invalid",
			},
			mockBehavior: func(s *MockAuthService, role string) {
				s.On("GenerateToken", role).Return("", apperrors.ErrInvalidRole)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Недопустимая роль\"}\n",
		},
		{
			name: "Empty Role",
			input: request.DummyLoginRequest{
				Role: "",
			},
			mockBehavior: func(s *MockAuthService, role string) {
				s.On("GenerateToken", role).Return("", apperrors.ErrInvalidRole)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Недопустимая роль\"}\n",
		},
		{
			name: "Service Error",
			input: request.DummyLoginRequest{
				Role: "employee",
			},
			mockBehavior: func(s *MockAuthService, role string) {
				s.On("GenerateToken", role).Return("", errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"message\":\"Внутренняя ошибка сервера\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockBehavior(mockService, tt.input.Role)
			handler := handlers.NewAuthHandler(mockService)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.DummyLogin(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

// Отдельные функции чтобы не ломать input в других тестах

func TestAuthHandler_DummyLogin_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	handler := handlers.NewAuthHandler(mockService)

	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer([]byte(`{"role": invalid_json`)))
	w := httptest.NewRecorder()

	handler.DummyLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"message\":\"Неверный формат запроса\"}\n", w.Body.String())
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	handler := handlers.NewAuthHandler(mockService)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte(`{"email": "test@example.com", "password": "password123", "role": invalid_json}`)))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"message\":\"Неверный формат запроса\"}\n", w.Body.String())
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	handler := handlers.NewAuthHandler(mockService)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte(`{"email": "test@example.com", "password": invalid_json}`)))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"message\":\"Неверный формат запроса\"}\n", w.Body.String())
}
