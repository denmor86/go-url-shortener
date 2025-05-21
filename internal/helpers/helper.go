package helpers

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const JWTExpire = time.Hour * 3

func MakeShortURL(urlValue string, size int) string {

	data := fmt.Sprintf("%s%d", urlValue, time.Now().UnixNano())

	hash := md5.Sum([]byte(data))

	return base64.URLEncoding.EncodeToString(hash[:])[:size]
}

func MakeURL(baseURL, shortURL string) string {

	var fullURL string
	fullURL = baseURL
	if !strings.HasSuffix(fullURL, "/") {
		fullURL += "/"
	}
	fullURL += shortURL
	return fullURL
}

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID string
}

func BuildJWT(userID string, secret []byte) (string, error) {
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
