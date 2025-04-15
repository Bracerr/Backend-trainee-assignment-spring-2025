package main

import (
	"fmt"
	"log"
	"net/http"

	"avito-backend/src/internal/config"
	"avito-backend/src/internal/delivery/http/handlers"
	"avito-backend/src/internal/delivery/http/routes"
	"avito-backend/src/internal/service"
	"avito-backend/src/pkg/jwt"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	tokenManager := jwt.NewTokenManager(cfg.JWTSigningKey, cfg.JWTTokenDuration)
	authService := service.NewAuthService(tokenManager)
	authHandler := handlers.NewAuthHandler(authService)

	router := routes.NewRouter(authHandler)

	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Сервер запущен на порту %s", cfg.ServerPort)
	if err := http.ListenAndServe(serverAddr, router.InitRoutes()); err != nil {
		log.Fatal(err)
	}
}
