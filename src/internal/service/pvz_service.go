package service

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/repository"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PVZServiceInterface interface {
	Create(city string) (*models.PVZ, error)
	CreateReception(pvzID uuid.UUID) (*models.Reception, error)
	CreateProduct(pvzID uuid.UUID, productType string) (*models.Product, error)
	DeleteLastProduct(pvzID uuid.UUID) error
	CloseLastReception(pvzID uuid.UUID) (*models.Reception, error)
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
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrPVZNotFound
	}
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

func (s *PVZService) CreateProduct(pvzID uuid.UUID, productType string) (*models.Product, error) {
	pType := models.ProductType(productType)
	if !pType.IsValid() {
		return nil, apperrors.ErrInvalidProductType
	}

	_, err := s.pvzRepo.GetByID(pvzID)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrPVZNotFound
	}
	if err != nil {
		return nil, err
	}

	activeReception, err := s.pvzRepo.GetActiveReceptionByPVZID(pvzID)
	if err != nil {
		return nil, err
	}
	if activeReception == nil {
		return nil, apperrors.ErrNoActiveReception
	}

	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        pType,
		ReceptionID: activeReception.ID,
	}

	if err := s.pvzRepo.CreateProduct(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *PVZService) DeleteLastProduct(pvzID uuid.UUID) error {
	_, err := s.pvzRepo.GetByID(pvzID)
	if err == sql.ErrNoRows {
		return apperrors.ErrPVZNotFound
	}
	if err != nil {
		return err
	}

	activeReception, err := s.pvzRepo.GetActiveReceptionByPVZID(pvzID)
	if err != nil {
		return err
	}
	if activeReception == nil {
		return apperrors.ErrNoActiveReception
	}
	if activeReception.Status == models.Closed {
		return apperrors.ErrReceptionClosed
	}

	lastProduct, err := s.pvzRepo.GetLastProductInReception(activeReception.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrNoProductsToDelete
		}
		return err
	}

	return s.pvzRepo.DeleteProduct(lastProduct.ID)
}

func (s *PVZService) CloseLastReception(pvzID uuid.UUID) (*models.Reception, error) {
	_, err := s.pvzRepo.GetByID(pvzID)
	if err == sql.ErrNoRows {
		return nil, apperrors.ErrPVZNotFound
	}
	if err != nil {
		return nil, err
	}

	reception, err := s.pvzRepo.GetActiveReceptionByPVZID(pvzID)
	if err != nil {
		return nil, err
	}
	if reception == nil {
		return nil, apperrors.ErrNoActiveReception
	}
	if reception.Status == models.Closed {
		return nil, apperrors.ErrReceptionAlreadyClosed
	}

	reception.Status = models.Closed
	err = s.pvzRepo.UpdateReception(reception)
	if err != nil {
		return nil, err
	}

	return reception, nil
}
