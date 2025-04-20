package service_test

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/service"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPVZRepository struct {
	mock.Mock
}

func (m *MockPVZRepository) Create(pvz *models.PVZ) error {
	args := m.Called(pvz)
	return args.Error(0)
}

func (m *MockPVZRepository) GetByID(id uuid.UUID) (*models.PVZ, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *MockPVZRepository) CreateReception(reception *models.Reception) error {
	args := m.Called(reception)
	return args.Error(0)
}

func (m *MockPVZRepository) GetActiveReceptionByPVZID(pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockPVZRepository) CreateProduct(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockPVZRepository) DeleteProduct(pvzID uuid.UUID) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func (m *MockPVZRepository) GetLastProductInReception(receptionID uuid.UUID) (*models.Product, error) {
	args := m.Called(receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockPVZRepository) UpdateReception(reception *models.Reception) error {
	args := m.Called(reception)
	return args.Error(0)
}

func (m *MockPVZRepository) GetPVZsWithReceptions(startDate, endDate time.Time, offset, limit int) ([]*models.PVZWithReceptions, error) {
	args := m.Called(startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PVZWithReceptions), args.Error(1)
}

func TestPVZService_Create(t *testing.T) {
	tests := []struct {
		name         string
		city         string
		mockBehavior func(repo *MockPVZRepository)
		wantErr      error
	}{
		{
			name: "Success Moscow",
			city: "Москва",
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("Create", mock.AnythingOfType("*models.PVZ")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Success SPB",
			city: "Санкт-Петербург",
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("Create", mock.AnythingOfType("*models.PVZ")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Success Kazan",
			city: "Казань",
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("Create", mock.AnythingOfType("*models.PVZ")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:         "Invalid City",
			city:         "Новосибирск",
			mockBehavior: func(repo *MockPVZRepository) {},
			wantErr:      apperrors.ErrInvalidCity,
		},
		{
			name:         "Empty City",
			city:         "",
			mockBehavior: func(repo *MockPVZRepository) {},
			wantErr:      apperrors.ErrInvalidCity,
		},
		{
			name: "Repository Error",
			city: "Москва",
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("Create", mock.AnythingOfType("*models.PVZ")).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockBehavior(mockRepo)
			service := service.NewPVZService(mockRepo)

			pvz, err := service.Create(tt.city)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, pvz)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pvz)
				assert.Equal(t, models.City(tt.city), pvz.City)
				assert.NotEmpty(t, pvz.ID)
				assert.NotZero(t, pvz.RegistrationDate)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_GetPVZsWithReceptions(t *testing.T) {
	tests := []struct {
		name         string
		startDate    time.Time
		endDate      time.Time
		offset       int
		limit        int
		mockBehavior func(repo *MockPVZRepository)
		wantErr      error
	}{
		{
			name:      "Success",
			startDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			offset:    0,
			limit:     10,
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetPVZsWithReceptions", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0, 10).Return(
					[]*models.PVZWithReceptions{
						{
							PVZ: &models.PVZ{
								ID:               uuid.New(),
								RegistrationDate: time.Now(),
								City:             models.City("Москва"),
							},
							Receptions: make([]models.ReceptionWithProducts, 0),
						},
					}, nil)
			},
			wantErr: nil,
		},
		{
			name:         "Invalid Date Range",
			startDate:    time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			endDate:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			offset:       0,
			limit:        10,
			mockBehavior: func(repo *MockPVZRepository) {},
			wantErr:      apperrors.ErrInvalidDateRange,
		},
		{
			name:         "Invalid Pagination",
			startDate:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:      time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			offset:       -1,
			limit:        10,
			mockBehavior: func(repo *MockPVZRepository) {},
			wantErr:      apperrors.ErrInvalidPagination,
		},
		{
			name:      "Repository Error",
			startDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			offset:    0,
			limit:     10,
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetPVZsWithReceptions", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0, 10).Return(
					nil, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockBehavior(mockRepo)
			service := service.NewPVZService(mockRepo)

			pvzs, err := service.GetPVZsWithReceptions(tt.startDate, tt.endDate, tt.offset, tt.limit)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, pvzs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pvzs)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
