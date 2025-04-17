package service

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

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
