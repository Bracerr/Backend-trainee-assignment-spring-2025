//go:build integration
package integration

import (
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/repository"
	"avito-backend/src/internal/service"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"avito-backend/src/pkg/database"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanupDatabase(db *sql.DB) error {
	tables := []string{"products", "receptions", "pvz"}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("ошибка при очистке таблицы %s: %w", table, err)
		}
	}
	return nil
}

func TestPVZFullCycle(t *testing.T) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_TEST_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB") + "_test",
	)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "../../../migrations"); err != nil {
		t.Fatalf("Ошибка применения миграций: %v", err)
	}

	if err := cleanupDatabase(db); err != nil {
		t.Fatalf("Ошибка при очистке базы данных: %v", err)
	}

	pvzRepo := repository.NewPVZRepository(db)
	pvzService := service.NewPVZService(pvzRepo)

	pvz, err := pvzService.Create(string(models.Moscow))
	require.NoError(t, err)
	require.NotNil(t, pvz)
	assert.Equal(t, models.Moscow, pvz.City)

	reception, err := pvzService.CreateReception(pvz.ID)
	require.NoError(t, err)
	require.NotNil(t, reception)
	assert.Equal(t, models.InProgress, reception.Status)
	assert.Equal(t, pvz.ID, reception.PVZID)

	productTypes := []models.ProductType{models.Electronics, models.Clothes, models.Shoes}
	for i := 0; i < 50; i++ {
		productType := productTypes[i%len(productTypes)]
		product, err := pvzService.CreateProduct(pvz.ID, string(productType))
		require.NoError(t, err)
		require.NotNil(t, product)
		assert.Equal(t, reception.ID, product.ReceptionID)
		assert.Equal(t, productType, product.Type)
	}

	closedReception, err := pvzService.CloseLastReception(pvz.ID)
	require.NoError(t, err)
	require.NotNil(t, closedReception)
	assert.Equal(t, models.Closed, closedReception.Status)
	assert.Equal(t, reception.ID, closedReception.ID)

	pvzs, err := pvzService.GetPVZsWithReceptions(time.Time{}, time.Time{}, 0, 10)
	require.NoError(t, err)
	require.Len(t, pvzs, 1)

	assert.Equal(t, pvz.ID, pvzs[0].PVZ.ID)
	assert.Equal(t, pvz.City, pvzs[0].PVZ.City)
	require.Len(t, pvzs[0].Receptions, 1)

	savedReception := pvzs[0].Receptions[0]
	assert.Equal(t, reception.ID, savedReception.Reception.ID)
	assert.Equal(t, models.Closed, savedReception.Reception.Status)
	assert.Len(t, savedReception.Products, 50)
}
