package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"avito-backend/src/internal/config"
	"avito-backend/src/internal/delivery/http/handlers"
	"avito-backend/src/internal/delivery/http/routes"
	"avito-backend/src/internal/repository"
	"avito-backend/src/internal/service"
	"avito-backend/src/pkg/database"
	"avito-backend/src/pkg/jwt"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	migrationsPath := filepath.Join("migrations")
	if err := database.RunMigrations(db, migrationsPath); err != nil {
		log.Printf("Ошибка применения миграций: %v", err)
	}

	tokenManager := jwt.NewTokenManager(cfg.JWTSigningKey, cfg.JWTTokenDuration)
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, tokenManager)
	authHandler := handlers.NewAuthHandler(authService)

	pvzRepo := repository.NewPVZRepository(db)
	pvzService := service.NewPVZService(pvzRepo)
	pvzHandler := handlers.NewPVZHandler(pvzService)

	router := routes.NewRouter(authHandler, pvzHandler, tokenManager)

	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Сервер запущен на порту %s", cfg.ServerPort)
	if err := http.ListenAndServe(serverAddr, router.InitRoutes()); err != nil {
		log.Fatal(err)
	}
}
