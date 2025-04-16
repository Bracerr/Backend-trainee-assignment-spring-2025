package routes

import (
	"avito-backend/src/internal/delivery/http/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Router struct {
	handler *handlers.AuthHandler
}

func NewRouter(handler *handlers.AuthHandler) *Router {
	return &Router{
		handler: handler,
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

	router.Post("/dummyLogin", r.handler.DummyLogin)
	router.Post("/register", r.handler.Register)
	router.Post("/login", r.handler.Login)

	return router
}
