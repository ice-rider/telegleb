package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type TokenManager struct {
	secretKey []byte
	ttl       time.Duration
}

func NewTokenManager(secretKey string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secretKey: []byte(secretKey),
		ttl:       ttl,
	}
}

func (tm *TokenManager) GenerateToken(sessionID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   sessionID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

func (tm *TokenManager) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return tm.secretKey, nil
	})

	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}