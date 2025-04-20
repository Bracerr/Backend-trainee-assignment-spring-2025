package jwt

import (
	"fmt"
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

func (m *TokenManager) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.signingKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.Role, nil
	}

	return "", fmt.Errorf("invalid token claims")
}
