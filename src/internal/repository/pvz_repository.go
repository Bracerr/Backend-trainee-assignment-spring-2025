package repository

import (
	"avito-backend/src/internal/domain/models"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type PVZRepositoryInterface interface {
	Create(pvz *models.PVZ) error
	GetByID(id uuid.UUID) (*models.PVZ, error)
}

type PVZRepository struct {
	db *sql.DB
}

func NewPVZRepository(db *sql.DB) *PVZRepository {
	return &PVZRepository{db: db}
}

func (r *PVZRepository) Create(pvz *models.PVZ) error {
	query := psql.Insert("pvz").
		Columns("id", "registration_date", "city").
		Values(pvz.ID, pvz.RegistrationDate, pvz.City)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(sql, args...)
	return err
}

func (r *PVZRepository) GetByID(id uuid.UUID) (*models.PVZ, error) {
	query := psql.Select("id", "registration_date", "city").
		From("pvz").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	pvz := &models.PVZ{}
	err = r.db.QueryRow(sql, args...).Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	if err != nil {
		return nil, err
	}

	return pvz, nil
}
