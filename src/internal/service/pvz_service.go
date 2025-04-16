package service

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/repository"
	"time"

	"github.com/google/uuid"
)

type PVZServiceInterface interface {
	Create(city string) (*models.PVZ, error)
}

type PVZService struct {
	pvzRepo repository.PVZRepositoryInterface
}

func NewPVZService(pvzRepo repository.PVZRepositoryInterface) PVZServiceInterface {
	return &PVZService{
		pvzRepo: pvzRepo,
	}
}

func (s *PVZService) Create(city string) (*models.PVZ, error) {
	cityEnum := models.City(city)
	if !cityEnum.IsValid() {
		return nil, apperrors.ErrInvalidCity
	}

	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             cityEnum,
	}

	if err := s.pvzRepo.Create(pvz); err != nil {
		return nil, err
	}

	return pvz, nil
}
