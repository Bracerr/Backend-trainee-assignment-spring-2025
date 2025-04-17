package service

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

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
		return err
	}
	if lastProduct == nil {
		return apperrors.ErrNoProductsToDelete
	}

	return s.pvzRepo.DeleteProduct(lastProduct.ID)
}
