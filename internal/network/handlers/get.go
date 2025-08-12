package handlers

import (
	"net/http"

	"github.com/denmor86/go-url-shortener/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

// DecodeURL - метод-обработчик получения запроса на получение оригинального URL по короткой ссылке
func DecodeURL(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var shortURL string
		id := chi.URLParam(r, "id")
		if id != "" {
			shortURL = id
		} else {
			// обратная совместимость
			shortURL = r.URL.Path[len("/"):]
		}

		url, err := u.DecodeURL(r.Context(), shortURL)
		if errors.Is(err, usecase.ErrDeletedViolation) {
			http.Error(w, errors.Cause(err).Error(), http.StatusGone)
			return
		}
		if err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(url))
	}
}

// PingStorage - метод-обработчик проверки соединения с хранилищем данных
func PingStorage(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := u.PingStorage(r.Context()); err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// GetURLS - метод-обработчик получения данных о сокращенных URL пользователя
func GetURLS(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if userID := r.Context().Value(usecase.UserIDContextKey); userID != nil {
			responce, err := u.GetURLS(r.Context(), userID.(string))
			if err != nil {
				http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
				return
			}
			if responce == nil {
				http.Error(w, "no user data", http.StatusNoContent)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responce)
		}
	}
}
