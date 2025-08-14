package handlers

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/denmor86/go-url-shortener/internal/usecase"
)

// EncondeURL - метод-обработчик получения запроса на формирования короткой ссылки
func EncondeURL(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userID := r.Context().Value(usecase.UserIDContextKey); userID != nil {
			shortURL, err := u.EncodeURL(r.Context(), r.Body, userID.(string))

			w.Header().Set("content-type", "text/plain")

			if err == nil {
				w.WriteHeader(http.StatusCreated)
				w.Write(shortURL)
				return
			}
			if errors.Is(err, usecase.ErrUniqueViolation) {
				w.WriteHeader(http.StatusConflict)
				w.Write(shortURL)
				return
			}

			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// EncondeURLJson - метод-обработчик получения запроса на формирование короткой ссылки. Тело запроса в формате JSON
func EncondeURLJson(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userID := r.Context().Value(usecase.UserIDContextKey); userID != nil {
			responce, err := u.EncodeURLJson(r.Context(), r.Body, userID.(string))

			w.Header().Set("Content-Type", "application/json")

			if err == nil {
				w.WriteHeader(http.StatusCreated)
				w.Write(responce)
				return
			}
			if errors.Is(err, usecase.ErrUniqueViolation) {
				w.WriteHeader(http.StatusConflict)
				w.Write(responce)
				return
			}

			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// EncondeURLJsonBatch - метод-обработчик получения запроса на формирование массива коротких ссылок. Тело запроса формате JSON
func EncondeURLJsonBatch(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userID := r.Context().Value(usecase.UserIDContextKey); userID != nil {
			responce, err := u.EncodeURLJsonBatch(r.Context(), r.Body, userID.(string))
			if err != nil {
				http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(responce)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
