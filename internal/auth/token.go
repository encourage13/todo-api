package auth

import (
	"fmt"
	"time"

	"github.com/encourage13/todo-api/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

func (t *TokenManager) Issue(userId uuid.UUID) (string, error) {
	issuedAt := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   userId.String(),
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(t.ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signedToken, nil
}

func (t *TokenManager) Parse(token string) (uuid.UUID, error) {
	var registeredClaims jwt.RegisteredClaims

	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.secret, nil
	}

	_, err := jwt.ParseWithClaims(token, &registeredClaims, keyFunc, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse token: %w, %w", model.ErrInvalidToken, err)
	}

	userIDStr := registeredClaims.Subject
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, model.ErrInvalidToken
	}
	return userID, nil
}
