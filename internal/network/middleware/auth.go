package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/denmor86/go-url-shortener.git/internal/usecase"
)

const (
	TokenCookie = "user-token"
)

type Authorization struct {
	Secret []byte
}

func NewAuthorization(cfg config.Config) *Authorization {
	return &Authorization{Secret: []byte(cfg.JWTSecret)}
}

func CheckCookie(secret []byte, r *http.Request) (string, error) {
	tokenCookie, err := r.Cookie(TokenCookie)
	if err != nil {
		// в запросе нет cookie
		return "", fmt.Errorf("the request does not contain cookies")
	}
	claims, err := helpers.ParseJWT(tokenCookie.Value, []byte(secret))
	if err != nil {
		// cookie не прошли валидацию
		return "", fmt.Errorf("invalid cookies")
	}
	if err = claims.Valid(); err != nil {
		return "", fmt.Errorf("invalid claims jwt")
	}

	return claims.UserID, nil
}

// CookieHandle — middleware-создание cookie для входящих HTTP-запросов.
func (auth *Authorization) CookieHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := CheckCookie(auth.Secret, r)
		if err != nil {
			// если coockie не прошли проверку, создаем новые
			userID = uuid.New().String()

			jwtToken, err := helpers.BuildJWT(userID, auth.Secret)
			if err != nil {
				h.ServeHTTP(w, r)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "user-token",
				MaxAge:   int(helpers.JWTExpire.Seconds()),
				HttpOnly: true,
				Value:    jwtToken,
			})
		}
		// создаем контекст, и добавляем в него ID пользователя (чтобы отвязать обработчик от парсинга cookie)
		ctx := context.WithValue(r.Context(), usecase.UserIDContextKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthHandle — middleware-аутентификация для входящих HTTP-запросов.
func (auth *Authorization) AuthHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie(TokenCookie); err != nil {
			// в запросе нет cookie, создаем новую
			userID := uuid.New().String()

			jwtToken, err := helpers.BuildJWT(userID, auth.Secret)
			if err != nil {
				h.ServeHTTP(w, r)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "user-token",
				MaxAge:   int(helpers.JWTExpire.Seconds()),
				HttpOnly: true,
				Value:    jwtToken,
			})
			w.WriteHeader(http.StatusNoContent)
			return
		}
		userID, err := CheckCookie(auth.Secret, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// создаем контекст, и добавляем в него ID пользователя (чтобы отвязать обработчик от парсинга cookie)
		ctx := context.WithValue(r.Context(), usecase.UserIDContextKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
