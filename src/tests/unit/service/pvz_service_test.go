package service_test

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/service"
	"database/sql"
	"errors"
	"testing"

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

func TestPVZService_CreateReception(t *testing.T) {
	tests := []struct {
		name         string
		pvzID        uuid.UUID
		mockBehavior func(repo *MockPVZRepository)
		wantErr      error
	}{
		{
			name:  "Success",
			pvzID: uuid.New(),
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&models.PVZ{
					ID:   uuid.New(),
					City: "Москва",
				}, nil)
				repo.On("GetActiveReceptionByPVZID", mock.AnythingOfType("uuid.UUID")).Return(nil, nil)
				repo.On("CreateReception", mock.AnythingOfType("*models.Reception")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "PVZ Not Found",
			pvzID: uuid.New(),
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(nil, sql.ErrNoRows)
			},
			wantErr: apperrors.ErrPVZNotFound,
		},
		{
			name:  "Active Reception Exists",
			pvzID: uuid.New(),
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&models.PVZ{
					ID:   uuid.New(),
					City: "Москва",
				}, nil)
				repo.On("GetActiveReceptionByPVZID", mock.AnythingOfType("uuid.UUID")).Return(&models.Reception{
					ID:     uuid.New(),
					Status: models.InProgress,
				}, nil)
			},
			wantErr: apperrors.ErrActiveReceptionExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockBehavior(mockRepo)
			service := service.NewPVZService(mockRepo)

			reception, err := service.CreateReception(tt.pvzID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, reception)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reception)
				assert.Equal(t, tt.pvzID, reception.PVZID)
				assert.Equal(t, models.InProgress, reception.Status)
				assert.NotEmpty(t, reception.ID)
				assert.NotZero(t, reception.DateTime)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
