package routes

import (
	"avito-backend/src/internal/delivery/http/handlers"
	appmiddleware "avito-backend/src/internal/delivery/http/middleware"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/pkg/jwt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Router struct {
	authHandler      *handlers.AuthHandler
	pvzHandler       *handlers.PVZHandler
	tokenManager     *jwt.TokenManager
}

func NewRouter(authHandler *handlers.AuthHandler, pvzHandler *handlers.PVZHandler, tokenManager *jwt.TokenManager) *Router {
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

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	router.Post("/dummyLogin", r.authHandler.DummyLogin)
	router.Post("/register", r.authHandler.Register)
	router.Post("/login", r.authHandler.Login)

	router.Group(func(router chi.Router) {
		router.Use(appmiddleware.AuthMiddleware(r.tokenManager))

		router.Group(func(router chi.Router) {
			router.Use(appmiddleware.RequireRole(models.ModeratorRole))
			router.Post("/pvz", r.pvzHandler.Create)
		})

		router.Group(func(router chi.Router) {
			router.Use(appmiddleware.RequireRole(models.EmployeeRole))
			router.Post("/receptions", r.pvzHandler.CreateReception)
		})
	})

	return router
}
