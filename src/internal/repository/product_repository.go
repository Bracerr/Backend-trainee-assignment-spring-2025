package repository

import (
	"avito-backend/src/internal/domain/models"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *PVZRepository) CreateProduct(product *models.Product) error {
	query := psql.Insert("products").
		Columns("id", "date_time", "type", "reception_id").
		Values(product.ID, product.DateTime, product.Type, product.ReceptionID)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(sqlQuery, args...)
	return err
}

func (r *PVZRepository) GetLastProductInReception(receptionID uuid.UUID) (*models.Product, error) {
	query := psql.Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(sq.Eq{"reception_id": receptionID}).
		OrderBy("date_time DESC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	product := &models.Product{}
	err = r.db.QueryRow(sqlQuery, args...).Scan(
		&product.ID,
		&product.DateTime,
		&product.Type,
		&product.ReceptionID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *PVZRepository) DeleteProduct(productID uuid.UUID) error {
	query := psql.Delete("products").
		Where(sq.Eq{"id": productID})

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
