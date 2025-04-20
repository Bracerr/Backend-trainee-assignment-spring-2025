package main

import (
	"database/sql"
	"log"
	"net"

	"avito-backend/src/internal/config"
	grpcdelivery "avito-backend/src/internal/delivery/grpc"
	"avito-backend/src/internal/delivery/grpc/pb"
	"avito-backend/src/internal/repository"
	"avito-backend/src/internal/service"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	pvzRepo := repository.NewPVZRepository(db)
	pvzService := service.NewPVZService(pvzRepo)

	grpcServer := grpc.NewServer()

	pb.RegisterPVZServiceServer(grpcServer, grpcdelivery.NewPVZGrpcServer(pvzService))

	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	log.Printf("gRPC сервер запущен на порту :3000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка при работе сервера: %v", err)
	}
}
