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
	CreateReception(pvzID uuid.UUID) (*models.Reception, error)
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

func (s *PVZService) CreateReception(pvzID uuid.UUID) (*models.Reception, error) {
	_, err := s.pvzRepo.GetByID(pvzID)
	if err != nil {
		return nil, err
	}

	activeReception, err := s.pvzRepo.GetActiveReceptionByPVZID(pvzID)
	if err != nil {
		return nil, err
	}
	if activeReception != nil {
		return nil, apperrors.ErrActiveReceptionExists
	}

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.InProgress,
	}

	if err := s.pvzRepo.CreateReception(reception); err != nil {
		return nil, err
	}

	return reception, nil
}
