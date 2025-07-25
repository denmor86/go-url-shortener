// Package helpers предоставляет функциональность вспомогательных функций приложения
package helpers

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims описание записей в токене JWT
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID string
}

// BuildJWT - метод для формирования JWT токена с добавлением UUID пользователя с использованием секрета
func BuildJWT(userID string, secret []byte) (string, error) {
	if len(secret) == 0 {
		return "", fmt.Errorf("empty secret")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWTExpire)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWT - метод разбора JWT токена с проверкой секрета и возвратом кастомных записей
func ParseJWT(tokenString string, secret []byte) (*JWTClaims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}
	return claims, nil
}
