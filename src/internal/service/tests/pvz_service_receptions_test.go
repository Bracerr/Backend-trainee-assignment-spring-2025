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
		{
			name:  "Error Getting Active Reception",
			pvzID: uuid.New(),
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&models.PVZ{
					ID:   uuid.New(),
					City: "Москва",
				}, nil)
				repo.On("GetActiveReceptionByPVZID", mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
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

func TestPVZService_CloseLastReception(t *testing.T) {
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
				repo.On("GetActiveReceptionByPVZID", mock.AnythingOfType("uuid.UUID")).Return(&models.Reception{
					ID:     uuid.New(),
					Status: models.InProgress,
				}, nil)
				repo.On("UpdateReception", mock.MatchedBy(func(r *models.Reception) bool {
					return r.Status == models.Closed
				})).Return(nil)
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
			name:  "No Active Reception",
			pvzID: uuid.New(),
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&models.PVZ{
					ID:   uuid.New(),
					City: "Москва",
				}, nil)
				repo.On("GetActiveReceptionByPVZID", mock.AnythingOfType("uuid.UUID")).Return(nil, nil)
			},
			wantErr: apperrors.ErrNoActiveReception,
		},
		{
			name:  "Reception Already Closed",
			pvzID: uuid.New(),
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&models.PVZ{
					ID:   uuid.New(),
					City: "Москва",
				}, nil)
				repo.On("GetActiveReceptionByPVZID", mock.AnythingOfType("uuid.UUID")).Return(&models.Reception{
					ID:     uuid.New(),
					Status: models.Closed,
				}, nil)
			},
			wantErr: apperrors.ErrReceptionAlreadyClosed,
		},
		{
			name:  "DB Error on GetByID",
			pvzID: uuid.New(),
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(nil, sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name:  "DB Error on GetActiveReceptionByPVZID",
			pvzID: uuid.New(),
			mockBehavior: func(repo *MockPVZRepository) {
				repo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&models.PVZ{
					ID:   uuid.New(),
					City: "Москва",
				}, nil)
				repo.On("GetActiveReceptionByPVZID", mock.AnythingOfType("uuid.UUID")).Return(nil, sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name:  "DB Error on UpdateReception",
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
				repo.On("UpdateReception", mock.MatchedBy(func(r *models.Reception) bool {
					return r.Status == models.Closed
				})).Return(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockBehavior(mockRepo)
			service := service.NewPVZService(mockRepo)

			reception, err := service.CloseLastReception(tt.pvzID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, reception)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reception)
				assert.Equal(t, models.Closed, reception.Status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
