package service

import (
	"avito-backend/src/internal/apperrors"
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/repository"
	"avito-backend/src/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	GenerateToken(role string) (string, error)
	Register(email, password, role string) (*models.User, error)
	Login(email, password string) (string, error)
}

type AuthService struct {
	userRepo     repository.UserRepositoryInterface
	tokenManager *jwt.TokenManager
}

func NewAuthService(userRepo repository.UserRepositoryInterface, tokenManager *jwt.TokenManager) AuthServiceInterface {
	return &AuthService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) validateRole(role string) error {
	if role != string(models.EmployeeRole) && role != string(models.ModeratorRole) {
		return apperrors.ErrInvalidRole
	}
	return nil
}

func (s *AuthService) Register(email, password, role string) (*models.User, error) {
	if err := s.validateRole(role); err != nil {
		return nil, err
	}

	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, apperrors.ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		Role:         role,
		PasswordHash: string(passwordHash),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", apperrors.ErrInvalidCredentials
	}

	return s.tokenManager.GenerateToken(user.Role)
}

// Выведено в отдельную функцию чтобы не тащить tokenManager в AuthHandler
func (s *AuthService) GenerateToken(role string) (string, error) {
	if err := s.validateRole(role); err != nil {
		return "", err
	}
	return s.tokenManager.GenerateToken(role)
}
