package repository_test

import (
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/repository"
	"regexp"
	"testing"
	"time"
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPVZRepository_GetPVZsWithReceptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewPVZRepository(db)

	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()
	now := time.Now()

	pvzQuery := regexp.QuoteMeta(`SELECT p.id, p.registration_date, p.city FROM pvz p LEFT JOIN receptions r ON p.id = r.pvz_id GROUP BY p.id, p.registration_date, p.city ORDER BY p.registration_date DESC LIMIT 10 OFFSET 0`)
	receptionQuery := regexp.QuoteMeta(`SELECT r.id, r.date_time, r.pvz_id, r.status FROM receptions r WHERE r.pvz_id = $1`)
	productQuery := regexp.QuoteMeta(`SELECT p.id, p.date_time, p.type, p.reception_id FROM products p WHERE p.reception_id = $1`)

	pvzRows := sqlmock.NewRows([]string{"id", "registration_date", "city"}).
		AddRow(pvzID, now, string(models.Moscow))

	receptionRows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
		AddRow(receptionID, now, pvzID, string(models.Closed))

	productRows := sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
		AddRow(productID, now, string(models.Electronics), receptionID)

	mock.ExpectQuery(pvzQuery).WillReturnRows(pvzRows)
	mock.ExpectQuery(receptionQuery).WithArgs(pvzID).WillReturnRows(receptionRows)
	mock.ExpectQuery(productQuery).WithArgs(receptionID).WillReturnRows(productRows)

	result, err := repo.GetPVZsWithReceptions(time.Time{}, time.Time{}, 0, 10)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Len(t, result, 1)

	assert.Equal(t, pvzID, result[0].PVZ.ID)
	assert.Equal(t, models.Moscow, result[0].PVZ.City)

	require.Len(t, result[0].Receptions, 1)
	assert.Equal(t, receptionID, result[0].Receptions[0].Reception.ID)
	assert.Equal(t, models.Closed, result[0].Receptions[0].Reception.Status)

	require.Len(t, result[0].Receptions[0].Products, 1)
	assert.Equal(t, productID, result[0].Receptions[0].Products[0].ID)
	assert.Equal(t, models.Electronics, result[0].Receptions[0].Products[0].Type)
}

func TestPVZRepository_GetPVZsWithReceptions_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewPVZRepository(db)

	pvzQuery := regexp.QuoteMeta(`SELECT p.id, p.registration_date, p.city FROM pvz p LEFT JOIN receptions r ON p.id = r.pvz_id GROUP BY p.id, p.registration_date, p.city ORDER BY p.registration_date DESC LIMIT 10 OFFSET 0`)

	mock.ExpectQuery(pvzQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}))

	result, err := repo.GetPVZsWithReceptions(time.Time{}, time.Time{}, 0, 10)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	assert.Len(t, result, 0)
}

func TestPVZRepository_GetPVZsWithReceptions_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewPVZRepository(db)

	pvzQuery := regexp.QuoteMeta(`SELECT p.id, p.registration_date, p.city FROM pvz p LEFT JOIN receptions r ON p.id = r.pvz_id GROUP BY p.id, p.registration_date, p.city ORDER BY p.registration_date DESC LIMIT 10 OFFSET 0`)

	mock.ExpectQuery(pvzQuery).WillReturnError(sql.ErrConnDone)

	result, err := repo.GetPVZsWithReceptions(time.Time{}, time.Time{}, 0, 10)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	assert.Nil(t, result)
}