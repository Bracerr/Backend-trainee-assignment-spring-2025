package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito-backend/src/internal/delivery/http/dto/request"
	"avito-backend/src/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_DummyLogin(t *testing.T) {
	tests := []struct {
		name         string
		input        request.DummyLoginRequest
		mockBehavior func(s *mocks.AuthServiceMock, role string)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success Employee",
			input: request.DummyLoginRequest{
				Role: "employee",
			},
			mockBehavior: func(s *mocks.AuthServiceMock, role string) {
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
			mockBehavior: func(s *mocks.AuthServiceMock, role string) {
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
			mockBehavior: func(s *mocks.AuthServiceMock, role string) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"message\":\"Недопустимая роль\"}\n",
		},
		{
			name: "Service Error",
			input: request.DummyLoginRequest{
				Role: "employee",
			},
			mockBehavior: func(s *mocks.AuthServiceMock, role string) {
				s.On("GenerateToken", role).Return("", errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"message\":\"service error\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.AuthServiceMock)
			tt.mockBehavior(mockService, tt.input.Role)
			handler := NewAuthHandler(mockService)

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

func TestAuthHandler_DummyLogin_InvalidJSON(t *testing.T) {
	mockService := new(mocks.AuthServiceMock)
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer([]byte(`{"role": invalid_json`)))
	w := httptest.NewRecorder()

	handler.DummyLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "{\"message\":\"Неверный формат запроса\"}\n", w.Body.String())
}
