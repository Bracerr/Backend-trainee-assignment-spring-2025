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
	CreateProduct(product *models.Product) error
	GetLastProductInReception(receptionID uuid.UUID) (*models.Product, error)
	DeleteProduct(productID uuid.UUID) error
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

	productsQuery := psql.Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(sq.Eq{"reception_id": reception.ID})

	sqlQuery, args, err = productsQuery.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err = rows.Scan(
			&product.ID,
			&product.DateTime,
			&product.Type,
			&product.ReceptionID,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	reception.Products = products
	return reception, nil
}

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
