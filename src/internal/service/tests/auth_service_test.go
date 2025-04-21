package service_test

import (
	"testing"

	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/repository"
	"avito-backend/src/internal/service"
	localjwt "avito-backend/src/pkg/jwt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := localjwt.NewTokenManager("test-secret", "24h")
	service := service.NewAuthService(mockRepo, tokenManager)

	tests := []struct {
		name         string
		email        string
		password     string
		role         string
		mockBehavior func(repo *MockUserRepository, user *models.User)
		wantErr      error
	}{
		{
			name:     "Success",
			email:    "test@example.com",
			password: "password123",
			role:     "employee",
			mockBehavior: func(repo *MockUserRepository, user *models.User) {
				repo.On("GetByEmail", "test@example.com").Return(nil, apperrors.ErrInvalidCredentials)
				repo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:     "User Already Exists",
			email:    "existing@example.com",
			password: "password123",
			role:     "employee",
			mockBehavior: func(repo *MockUserRepository, user *models.User) {
				repo.On("GetByEmail", "existing@example.com").Return(&models.User{}, nil)
			},
			wantErr: apperrors.ErrUserAlreadyExists,
		},
		{
			name:     "Invalid Role",
			email:    "test@example.com",
			password: "password123",
			role:     "invalid_role",
			mockBehavior: func(repo *MockUserRepository, user *models.User) {
			},
			wantErr: apperrors.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockBehavior(mockRepo, nil)

			user, err := service.Register(tt.email, tt.password, tt.role)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.role, user.Role)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenManager := localjwt.NewTokenManager("test-secret", "24h")
	service := service.NewAuthService(mockRepo, tokenManager)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name         string
		email        string
		password     string
		mockBehavior func(repo *MockUserRepository)
		wantErr      error
	}{
		{
			name:     "Success Employee",
			email:    "test@example.com",
			password: password,
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("GetByEmail", "test@example.com").Return(&models.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "employee",
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:     "Success Moderator",
			email:    "mod@example.com",
			password: password,
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("GetByEmail", "mod@example.com").Return(&models.User{
					ID:           uuid.New(),
					Email:        "mod@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "moderator",
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:     "User Not Found",
			email:    "nonexistent@example.com",
			password: "password123",
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("GetByEmail", "nonexistent@example.com").Return(nil, apperrors.ErrInvalidCredentials)
			},
			wantErr: apperrors.ErrInvalidCredentials,
		},
		{
			name:     "Wrong Password",
			email:    "test@example.com",
			password: "wrongpassword",
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("GetByEmail", "test@example.com").Return(&models.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "employee",
				}, nil)
			},
			wantErr: apperrors.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockBehavior(mockRepo)

			token, err := service.Login(tt.email, tt.password)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				claims := jwt.MapClaims{}
				_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret"), nil
				})
				assert.NoError(t, err)
				expectedRole := "employee"
				if tt.name == "Success Moderator" {
					expectedRole = "moderator"
				}
				assert.Equal(t, expectedRole, claims["role"])
			}
			if tt.name != "Empty Email" && tt.name != "Empty Password" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	tokenManager := localjwt.NewTokenManager("test-secret", "24h")
	userRepo := &repository.UserRepository{}
	service := service.NewAuthService(userRepo, tokenManager)

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{
			name:    "Success Employee",
			role:    "employee",
			wantErr: false,
		},
		{
			name:    "Success Moderator",
			role:    "moderator",
			wantErr: false,
		},
		{
			name:    "Invalid Role",
			role:    "invalid_role",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				claims := jwt.MapClaims{}
				_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret"), nil
				})
				assert.NoError(t, err)
				assert.Equal(t, tt.role, claims["role"])
			}
		})
	}
}
