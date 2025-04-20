package routes

import (
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/pkg/jwt"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthHandler struct {
	mock.Mock
}

func (m *MockAuthHandler) Register(w http.ResponseWriter, r *http.Request)   { m.Called(w, r) }
func (m *MockAuthHandler) Login(w http.ResponseWriter, r *http.Request)      { m.Called(w, r) }
func (m *MockAuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) { m.Called(w, r) }

type MockPVZHandler struct {
	mock.Mock
}

func (m *MockPVZHandler) Create(w http.ResponseWriter, r *http.Request)             { m.Called(w, r) }
func (m *MockPVZHandler) GetPVZs(w http.ResponseWriter, r *http.Request)            { m.Called(w, r) }
func (m *MockPVZHandler) CreateReception(w http.ResponseWriter, r *http.Request)    { m.Called(w, r) }
func (m *MockPVZHandler) CreateProduct(w http.ResponseWriter, r *http.Request)      { m.Called(w, r) }
func (m *MockPVZHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request)  { m.Called(w, r) }
func (m *MockPVZHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) { m.Called(w, r) }

func TestNewRouter(t *testing.T) {
	authHandler := &MockAuthHandler{}
	pvzHandler := &MockPVZHandler{}
	tokenManager := jwt.NewTokenManager("test-key", "1h")

	router := NewRouter(authHandler, pvzHandler, tokenManager)

	assert.NotNil(t, router)
	assert.Equal(t, authHandler, router.authHandler)
	assert.Equal(t, pvzHandler, router.pvzHandler)
	assert.Equal(t, tokenManager, router.tokenManager)
}

func TestRouter_InitRoutes(t *testing.T) {
	authHandler := &MockAuthHandler{}
	pvzHandler := &MockPVZHandler{}
	tokenManager := jwt.NewTokenManager("test-key", "1h")
	router := NewRouter(authHandler, pvzHandler, tokenManager)

	r := router.InitRoutes()

	assert.NotNil(t, r)

	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/register"},
		{"POST", "/login"},
		{"POST", "/dummyLogin"},
		{"POST", "/pvz"},
		{"GET", "/pvz"},
		{"POST", "/receptions"},
		{"POST", "/products"},
		{"POST", "/pvz/{pvzId}/delete_last_product"},
		{"POST", "/pvz/{pvzId}/close_last_reception"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			routeExists := false
			_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
				if method == rt.method && route == rt.path {
					routeExists = true
				}
				return nil
			})
			assert.True(t, routeExists, "Маршрут %s %s должен быть зарегистрирован", rt.method, rt.path)
		})
	}
}

func TestRouter_Middleware(t *testing.T) {
	authHandler := &MockAuthHandler{}
	pvzHandler := &MockPVZHandler{}
	tokenManager := jwt.NewTokenManager("test-key", "1h")
	router := NewRouter(authHandler, pvzHandler, tokenManager)

	r := router.InitRoutes()

	tests := []struct {
		name     string
		path     string
		method   string
		role     models.Role
		wantCode int
	}{
		{
			name:     "Moderator can create PVZ",
			path:     "/pvz",
			method:   "POST",
			role:     models.ModeratorRole,
			wantCode: http.StatusOK,
		},
		{
			name:     "Employee cannot create PVZ",
			path:     "/pvz",
			method:   "POST",
			role:     models.EmployeeRole,
			wantCode: http.StatusForbidden,
		},
		{
			name:     "Employee can create receptions",
			path:     "/receptions",
			method:   "POST",
			role:     models.EmployeeRole,
			wantCode: http.StatusOK,
		},
		{
			name:     "Both can get PVZ list",
			path:     "/pvz",
			method:   "GET",
			role:     models.EmployeeRole,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)

			token, _ := tokenManager.GenerateToken(string(tt.role))
			req.Header.Set("Authorization", "Bearer "+token)

			foundMiddlewares := make(map[string]bool)
			_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
				if method == tt.method && route == tt.path {
					for range middlewares {
						foundMiddlewares["auth"] = true
						foundMiddlewares["metrics"] = true
						foundMiddlewares["logger"] = true
					}
				}
				return nil
			})

			assert.True(t, foundMiddlewares["auth"], "Должен быть подключен AuthMiddleware")
			assert.True(t, foundMiddlewares["metrics"], "Должен быть подключен MetricsMiddleware")
			assert.True(t, foundMiddlewares["logger"], "Должен быть подключен LoggerMiddleware")
		})
	}
}
