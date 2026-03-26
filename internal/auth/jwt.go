package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

const tokenExpiry = 24 * 30 * time.Hour // 30 days

type claims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
}

func NewToken(secret, userID string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
		},
		UserID: userID,
	})
	return t.SignedString(secretBytes(secret))
}

func ParseToken(secret, tokenString string) (string, error) {
	t, err := jwt.ParseWithClaims(tokenString, &claims{}, func(_ *jwt.Token) (interface{}, error) {
		return secretBytes(secret), nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}
	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid || c.UserID == "" {
		return "", ErrInvalidToken
	}
	return c.UserID, nil
}

func secretBytes(s string) []byte {
	if s == "" {
		return []byte("default-secret")
	}
	return []byte(s)
}
