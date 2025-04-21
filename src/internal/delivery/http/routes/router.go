package routes

import (
	appmiddleware "avito-backend/src/internal/delivery/http/middleware"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/pkg/jwt"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type AuthHandlerInterface interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	DummyLogin(w http.ResponseWriter, r *http.Request)
}

type PVZHandlerInterface interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetPVZs(w http.ResponseWriter, r *http.Request)
	CreateReception(w http.ResponseWriter, r *http.Request)
	CreateProduct(w http.ResponseWriter, r *http.Request)
	DeleteLastProduct(w http.ResponseWriter, r *http.Request)
	CloseLastReception(w http.ResponseWriter, r *http.Request)
}

type Router struct {
	authHandler  AuthHandlerInterface
	pvzHandler   PVZHandlerInterface
	tokenManager *jwt.TokenManager
}

func NewRouter(authHandler AuthHandlerInterface, pvzHandler PVZHandlerInterface, tokenManager *jwt.TokenManager) *Router {
	return &Router{
		authHandler:  authHandler,
		pvzHandler:   pvzHandler,
		tokenManager: tokenManager,
	}
}

func (r *Router) InitRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	router.Use(appmiddleware.MetricsMiddleware)
	router.Use(middleware.RequestID)
	router.Use(appmiddleware.LoggerMiddleware)

	router.Post("/register", r.authHandler.Register)
	router.Post("/login", r.authHandler.Login)
	router.Post("/dummyLogin", r.authHandler.DummyLogin)

	router.Group(func(router chi.Router) {
		router.Use(appmiddleware.AuthMiddleware(r.tokenManager))

		router.Group(func(router chi.Router) {
			router.Use(appmiddleware.RequireRole(models.ModeratorRole))
			router.Post("/pvz", r.pvzHandler.Create)
		})

		router.Group(func(router chi.Router) {
			router.Use(appmiddleware.RequireRole(models.EmployeeRole))
			router.Post("/receptions", r.pvzHandler.CreateReception)
			router.Post("/products", r.pvzHandler.CreateProduct)
			router.Post("/pvz/{pvzId}/delete_last_product", r.pvzHandler.DeleteLastProduct)
			router.Post("/pvz/{pvzId}/close_last_reception", r.pvzHandler.CloseLastReception)
		})

		router.Group(func(router chi.Router) {
			router.Use(appmiddleware.RequireRoles([]models.Role{models.EmployeeRole, models.ModeratorRole}))
			router.Get("/pvz", r.pvzHandler.GetPVZs)
		})
	})

	return router
}
