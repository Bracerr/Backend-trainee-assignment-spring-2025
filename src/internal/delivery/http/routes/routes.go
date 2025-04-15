package routes

import (
	"avito-backend/src/internal/delivery/http/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	router.Group(func(router chi.Router) {
		router.Post("/dummyLogin", r.handler.DummyLogin)
	})

	return router
}
