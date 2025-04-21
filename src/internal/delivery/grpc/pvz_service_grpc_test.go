package grpc_test

import (
	"avito-backend/src/internal/delivery/grpc"
	pb "avito-backend/src/internal/delivery/grpc/pb"
	"avito-backend/src/internal/domain/models"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPVZService struct {
	mock.Mock
}

func (m *MockPVZService) GetPVZsWithReceptions(startDate, endDate time.Time, offset, limit int) ([]*models.PVZWithReceptions, error) {
	args := m.Called(startDate, endDate, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PVZWithReceptions), args.Error(1)
}

func (m *MockPVZService) Create(city string) (*models.PVZ, error) {
	return nil, nil
}

func (m *MockPVZService) CreateReception(pvzID uuid.UUID) (*models.Reception, error) {
	return nil, nil
}

func (m *MockPVZService) CreateProduct(pvzID uuid.UUID, productType string) (*models.Product, error) {
	return nil, nil
}

func (m *MockPVZService) CloseLastReception(pvzID uuid.UUID) (*models.Reception, error) {
	return nil, nil
}

func (m *MockPVZService) DeleteLastProduct(pvzID uuid.UUID) error {
	return nil
}

func TestPVZGrpcServer_GetPVZList(t *testing.T) {
	tests := []struct {
		name         string
		mockBehavior func(s *MockPVZService)
		request      *pb.GetPVZListRequest
		checkResult  func(t *testing.T, response *pb.GetPVZListResponse, err error)
	}{
		{
			name: "Success",
			mockBehavior: func(s *MockPVZService) {
				pvzID := uuid.New()
				now := time.Now()
				pvzList := []*models.PVZWithReceptions{
					{
						PVZ: &models.PVZ{
							ID:               pvzID,
							RegistrationDate: now,
							City:             models.Moscow,
						},
						Receptions: []models.ReceptionWithProducts{},
					},
				}
				s.On("GetPVZsWithReceptions",
					time.Time{},
					time.Time{},
					0,
					1000,
				).Return(pvzList, nil)
			},
			request: &pb.GetPVZListRequest{},
			checkResult: func(t *testing.T, response *pb.GetPVZListResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Len(t, response.Pvzs, 1)
				assert.Equal(t, string(models.Moscow), response.Pvzs[0].City)
				assert.NotEmpty(t, response.Pvzs[0].Id)
				assert.NotNil(t, response.Pvzs[0].RegistrationDate)
			},
		},
		{
			name: "Empty Result",
			mockBehavior: func(s *MockPVZService) {
				s.On("GetPVZsWithReceptions",
					time.Time{},
					time.Time{},
					0,
					1000,
				).Return([]*models.PVZWithReceptions{}, nil)
			},
			request: &pb.GetPVZListRequest{},
			checkResult: func(t *testing.T, response *pb.GetPVZListResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Empty(t, response.Pvzs)
			},
		},
		{
			name: "Service Error",
			mockBehavior: func(s *MockPVZService) {
				s.On("GetPVZsWithReceptions",
					time.Time{},
					time.Time{},
					0,
					1000,
				).Return(nil, sql.ErrConnDone)
			},
			request: &pb.GetPVZListRequest{},
			checkResult: func(t *testing.T, response *pb.GetPVZListResponse, err error) {
				require.Error(t, err)
				assert.Equal(t, sql.ErrConnDone, err)
				assert.Nil(t, response)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.mockBehavior(mockService)

			server := grpc.NewPVZGrpcServer(mockService)

			response, err := server.GetPVZList(context.Background(), tt.request)

			tt.checkResult(t, response, err)

			mockService.AssertExpectations(t)
		})
	}
}
