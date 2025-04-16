package service_test

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/service"
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
