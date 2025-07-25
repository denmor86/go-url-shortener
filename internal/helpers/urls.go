// Package helpers предоставляет функциональность вспомогательных функций приложения
package helpers

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// JWTExpire - время жизни токена
const JWTExpire = time.Hour * 3

// MakeShortURL - метод формирования короткого URL c использованием базового URL и длинны необходимого короткого URL
func MakeShortURL(urlValue string, size int) (string, error) {
	// Проверка корректности размера
	if err := checkMD5HashSize(size); err != nil {
		return "", fmt.Errorf("invalid size: %w", err)
	}
	if urlValue == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	data := fmt.Sprintf("%s%d", urlValue, time.Now().UnixNano())

	hash := md5.Sum([]byte(data))

	encoded := base64.URLEncoding.EncodeToString(hash[:])

	return encoded[:size], nil
}

// MakeURL - метод формирования полного URL на основе базового URL и короткой ссылки
func MakeURL(baseURL, shortURL string) string {

	var fullURL string
	fullURL = baseURL
	if !strings.HasSuffix(fullURL, "/") {
		fullURL += "/"
	}
	fullURL += shortURL
	return fullURL
}

// checkMD5HashSize - метод проверки корректности размера генерируемой короткой ссылки
func checkMD5HashSize(size int) error {
	// Проверка корректности размера
	if size <= 0 {
		return fmt.Errorf("size must be positive, got %d", size)
	}
	if size > 24 { // Максимальная длина для base64 от MD5 (24 символа)
		return fmt.Errorf("size too large, maximum is 24, got %d", size)
	}
	return nil
}
