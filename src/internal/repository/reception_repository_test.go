package repository_test

import (
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/repository"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPVZRepository_CreateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewPVZRepository(db)
	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	reception := &models.Reception{
		ID:       receptionID,
		DateTime: now,
		PVZID:    pvzID,
		Status:   models.InProgress,
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO receptions (id,date_time,pvz_id,status) VALUES ($1,$2,$3,$4)`)).
		WithArgs(reception.ID, reception.DateTime, reception.PVZID, reception.Status).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateReception(reception)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPVZRepository_GetActiveReceptionByPVZID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewPVZRepository(db)
	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, now, pvzID, models.InProgress)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT r.id, r.date_time, r.pvz_id, r.status FROM receptions r WHERE (r.pvz_id = $1 AND r.status = $2)`)).
			WithArgs(pvzID, models.InProgress).
			WillReturnRows(rows)

		reception, err := repo.GetActiveReceptionByPVZID(pvzID)
		require.NoError(t, err)
		require.NotNil(t, reception)
		assert.Equal(t, receptionID, reception.ID)
		assert.Equal(t, pvzID, reception.PVZID)
		assert.Equal(t, models.InProgress, reception.Status)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT r.id, r.date_time, r.pvz_id, r.status FROM receptions r WHERE (r.pvz_id = $1 AND r.status = $2)`)).
			WithArgs(pvzID, models.InProgress).
			WillReturnError(sql.ErrNoRows)

		reception, err := repo.GetActiveReceptionByPVZID(pvzID)
		require.NoError(t, err)
		assert.Nil(t, reception)
	})

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT r.id, r.date_time, r.pvz_id, r.status FROM receptions r WHERE (r.pvz_id = $1 AND r.status = $2)`)).
			WithArgs(pvzID, models.InProgress).
			WillReturnError(sql.ErrConnDone)

		reception, err := repo.GetActiveReceptionByPVZID(pvzID)
		require.Error(t, err)
		assert.Nil(t, reception)
	})
}

func TestPVZRepository_UpdateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewPVZRepository(db)
	receptionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		reception := &models.Reception{
			ID:     receptionID,
			Status: models.Closed,
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE receptions SET status = $1 WHERE id = $2`)).
			WithArgs(reception.Status, reception.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.UpdateReception(reception)
		require.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		reception := &models.Reception{
			ID:     receptionID,
			Status: models.Closed,
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE receptions SET status = $1 WHERE id = $2`)).
			WithArgs(reception.Status, reception.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = repo.UpdateReception(reception)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		reception := &models.Reception{
			ID:     receptionID,
			Status: models.Closed,
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE receptions SET status = $1 WHERE id = $2`)).
			WithArgs(reception.Status, reception.ID).
			WillReturnError(sql.ErrConnDone)

		err = repo.UpdateReception(reception)
		require.Error(t, err)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
