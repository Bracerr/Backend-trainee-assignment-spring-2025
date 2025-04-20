package repository

import (
	"avito-backend/src/internal/domain/models"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

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
	query := psql.Select("r.id", "r.date_time", "r.pvz_id", "r.status").
		From("receptions r").
		Where(sq.And{
			sq.Eq{"r.pvz_id": pvzID},
			sq.Eq{"r.status": models.InProgress},
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

func (r *PVZRepository) UpdateReception(reception *models.Reception) error {
	query := psql.Update("receptions").
		Set("status", reception.Status).
		Where(sq.Eq{"id": reception.ID})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.Exec(sqlQuery, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
