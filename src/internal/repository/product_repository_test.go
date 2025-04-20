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
func TestPVZRepository_CreateProduct(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := repository.NewPVZRepository(db)
    productID := uuid.New()
    receptionID := uuid.New()
    now := time.Now()

    product := &models.Product{
        ID:          productID,
        DateTime:    now,
        Type:        models.Electronics,
        ReceptionID: receptionID,
    }

    mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO products (id,date_time,type,reception_id) VALUES ($1,$2,$3,$4)`)).
        WithArgs(product.ID, product.DateTime, product.Type, product.ReceptionID).
        WillReturnResult(sqlmock.NewResult(1, 1))

    err = repo.CreateProduct(product)
    require.NoError(t, err)
    require.NoError(t, mock.ExpectationsWereMet())
}

func TestPVZRepository_GetLastProductInReception(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := repository.NewPVZRepository(db)
    productID := uuid.New()
    receptionID := uuid.New()
    now := time.Now()

    t.Run("Success", func(t *testing.T) {
        rows := sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
            AddRow(productID, now, models.Electronics, receptionID)

        mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, date_time, type, reception_id FROM products WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1`)).
            WithArgs(receptionID).
            WillReturnRows(rows)

        product, err := repo.GetLastProductInReception(receptionID)
        require.NoError(t, err)
        require.NotNil(t, product)
        assert.Equal(t, productID, product.ID)
        assert.Equal(t, models.Electronics, product.Type)
        assert.Equal(t, receptionID, product.ReceptionID)
    })

    t.Run("Not Found", func(t *testing.T) {
        mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, date_time, type, reception_id FROM products WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1`)).
            WithArgs(receptionID).
            WillReturnError(sql.ErrNoRows)

        product, err := repo.GetLastProductInReception(receptionID)
        require.NoError(t, err)
        assert.Nil(t, product)
    })

    t.Run("DB Error", func(t *testing.T) {
        mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, date_time, type, reception_id FROM products WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1`)).
            WithArgs(receptionID).
            WillReturnError(sql.ErrConnDone)

        product, err := repo.GetLastProductInReception(receptionID)
        require.Error(t, err)
        assert.Nil(t, product)
    })
}

func TestPVZRepository_DeleteProduct(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := repository.NewPVZRepository(db)
    productID := uuid.New()

    t.Run("Success", func(t *testing.T) {
        mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM products WHERE id = $1`)).
            WithArgs(productID).
            WillReturnResult(sqlmock.NewResult(1, 1))

        err = repo.DeleteProduct(productID)
        require.NoError(t, err)
    })

    t.Run("Not Found", func(t *testing.T) {
        mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM products WHERE id = $1`)).
            WithArgs(productID).
            WillReturnResult(sqlmock.NewResult(0, 0))

        err = repo.DeleteProduct(productID)
        assert.Equal(t, sql.ErrNoRows, err)
    })

    t.Run("DB Error", func(t *testing.T) {
        mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM products WHERE id = $1`)).
            WithArgs(productID).
            WillReturnError(sql.ErrConnDone)

        err = repo.DeleteProduct(productID)
        require.Error(t, err)
    })

    require.NoError(t, mock.ExpectationsWereMet())
}