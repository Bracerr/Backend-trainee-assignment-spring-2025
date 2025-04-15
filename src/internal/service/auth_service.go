package service

import (
	"avito-backend/src/pkg/jwt"
)

type AuthServiceInterface interface {
	GenerateToken(role string) (string, error)
}

type AuthService struct {
	tokenManager *jwt.TokenManager
}

func NewAuthService(tokenManager *jwt.TokenManager) AuthServiceInterface {
	return &AuthService{
		tokenManager: tokenManager,
	}
}

func (s *AuthService) GenerateToken(role string) (string, error) {
	return s.tokenManager.GenerateToken(role)
}
