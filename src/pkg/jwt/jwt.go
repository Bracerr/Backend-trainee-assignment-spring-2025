package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	signingKey string
	duration   string
}

func NewTokenManager(signingKey string, duration string) *TokenManager {
	return &TokenManager{
		signingKey: signingKey,
		duration:   duration,
	}
}

func (m *TokenManager) GenerateToken(role string) (string, error) {
	duration, err := time.ParseDuration(m.duration)
	if err != nil {
		return "", err
	}

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.signingKey))
}
