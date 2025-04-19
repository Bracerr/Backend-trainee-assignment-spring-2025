package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"avito-backend/src/internal/config"
	"avito-backend/src/internal/delivery/http/handlers"
	"avito-backend/src/internal/delivery/http/routes"
	"avito-backend/src/internal/repository"
	"avito-backend/src/internal/service"
	"avito-backend/src/pkg/database"
	"avito-backend/src/pkg/jwt"
	"avito-backend/src/pkg/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.InitLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

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

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.MetricsPort),
		Handler: promhttp.Handler(),
	}

	mainServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: router.InitRoutes(),
	}

	go func() {
		log.Printf("Сервер метрик запущен на порту :9000")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка сервера метрик: %v", err)
		}
	}()

	go func() {
		log.Printf("Основной сервер запущен на порту %s", cfg.ServerPort)
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка основного сервера: %v", err)
		}
	}()

	<-quit
	log.Println("Начинаем завершение работы серверов...")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Принудительное завершение сервера метрик: %v", err)
	}

	if err := mainServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Принудительное завершение основного сервера: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("Ошибка закрытия соединения с БД: %v", err)
	}

	log.Println("Серверы успешно остановлены")
}
