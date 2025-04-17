package repository

import (
	"avito-backend/src/internal/domain/models"
	"database/sql"
	"time"

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
	UpdateReception(reception *models.Reception) error
	GetPVZsWithReceptions(startDate, endDate time.Time, offset, limit int) ([]*models.PVZWithReceptions, error)
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

func (r *PVZRepository) GetPVZsWithReceptions(startDate, endDate time.Time, offset, limit int) ([]*models.PVZWithReceptions, error) {
	query := psql.Select("p.id", "p.registration_date", "p.city").
		From("pvz p").
		LeftJoin("receptions r ON p.id = r.pvz_id")

	if !startDate.IsZero() && !endDate.IsZero() {
		query = query.Where(sq.And{
			sq.GtOrEq{"r.date_time": startDate},
			sq.LtOrEq{"r.date_time": endDate},
		})
	}

	query = query.GroupBy("p.id", "p.registration_date", "p.city").
		OrderBy("p.registration_date DESC").
		Offset(uint64(offset)).
		Limit(uint64(limit))

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pvzs []*models.PVZWithReceptions
	for rows.Next() {
		pvz := &models.PVZ{}
		err = rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
		if err != nil {
			return nil, err
		}

		pvzWithReceptions := &models.PVZWithReceptions{
			PVZ:        pvz,
			Receptions: make([]models.ReceptionWithProducts, 0),
		}

		receptionsQuery := psql.Select("r.id", "r.date_time", "r.pvz_id", "r.status").
			From("receptions r").
			Where(sq.Eq{"r.pvz_id": pvz.ID})

		if !startDate.IsZero() && !endDate.IsZero() {
			receptionsQuery = receptionsQuery.Where(sq.And{
				sq.GtOrEq{"r.date_time": startDate},
				sq.LtOrEq{"r.date_time": endDate},
			})
		}

		sqlQuery, args, err = receptionsQuery.ToSql()
		if err != nil {
			return nil, err
		}

		receptionRows, err := r.db.Query(sqlQuery, args...)
		if err != nil {
			return nil, err
		}

		for receptionRows.Next() {
			reception := &models.Reception{}
			err = receptionRows.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status)
			if err != nil {
				receptionRows.Close()
				return nil, err
			}

			receptionWithProducts := models.ReceptionWithProducts{
				Reception: reception,
				Products:  make([]models.Product, 0),
			}

			productsQuery := psql.Select("p.id", "p.date_time", "p.type", "p.reception_id").
				From("products p").
				Where(sq.Eq{"p.reception_id": reception.ID})

			sqlQuery, args, err = productsQuery.ToSql()
			if err != nil {
				receptionRows.Close()
				return nil, err
			}

			productRows, err := r.db.Query(sqlQuery, args...)
			if err != nil {
				receptionRows.Close()
				return nil, err
			}

			for productRows.Next() {
				product := models.Product{}
				err = productRows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID)
				if err != nil {
					productRows.Close()
					receptionRows.Close()
					return nil, err
				}
				receptionWithProducts.Products = append(receptionWithProducts.Products, product)
			}
			productRows.Close()

			pvzWithReceptions.Receptions = append(pvzWithReceptions.Receptions, receptionWithProducts)
		}
		receptionRows.Close()

		pvzs = append(pvzs, pvzWithReceptions)
	}

	if pvzs == nil {
		pvzs = make([]*models.PVZWithReceptions, 0)
	}

	return pvzs, nil
}
