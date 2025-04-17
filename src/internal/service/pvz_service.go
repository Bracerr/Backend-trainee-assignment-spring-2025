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
	CreateProduct(pvzID uuid.UUID, productType string) (*models.Product, error)
	DeleteLastProduct(pvzID uuid.UUID) error
	CloseLastReception(pvzID uuid.UUID) (*models.Reception, error)
	GetPVZsWithReceptions(startDate, endDate time.Time, offset, limit int) ([]*models.PVZWithReceptions, error)
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

func (s *PVZService) GetPVZsWithReceptions(startDate, endDate time.Time, offset, limit int) ([]*models.PVZWithReceptions, error) {
	if !startDate.IsZero() && !endDate.IsZero() && startDate.After(endDate) {
		return nil, apperrors.ErrInvalidDateRange
	}

	if offset < 0 || limit <= 0 {
		return nil, apperrors.ErrInvalidPagination
	}

	pvzs, err := s.pvzRepo.GetPVZsWithReceptions(startDate, endDate, offset, limit)
	if err != nil {
		return nil, err
	}

	return pvzs, nil
}
