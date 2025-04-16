package repository

import (
	"avito-backend/src/internal/domain/models"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type PVZRepositoryInterface interface {
	Create(pvz *models.PVZ) error
	GetByID(id uuid.UUID) (*models.PVZ, error)
	CreateReception(reception *models.Reception) error
	GetActiveReceptionByPVZID(pvzID uuid.UUID) (*models.Reception, error)
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
		Where(sq.Eq{"id": id})

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

func (r *PVZRepository) CreateReception(reception *models.Reception) error {
	query := psql.Insert("receptions").
		Columns("id", "date_time", "pvz_id", "status").
		Values(reception.ID, reception.DateTime, reception.PVZID, reception.Status)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(sqlQuery, args...)
	return err
}

func (r *PVZRepository) GetActiveReceptionByPVZID(pvzID uuid.UUID) (*models.Reception, error) {
	query := psql.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(sq.And{
			sq.Eq{"pvz_id": pvzID},
			sq.Eq{"status": models.InProgress},
		})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	reception := &models.Reception{}
	err = r.db.QueryRow(sqlQuery, args...).Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PVZID,
		&reception.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return reception, nil
}
