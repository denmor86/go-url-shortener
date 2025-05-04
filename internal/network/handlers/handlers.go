package handlers

import (
	"net/http"
	"time"

	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

func EncondeURL(baseURL string, lenShortURL int, storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		shortURL, err := usecase.EncondeURL(r.Context(), baseURL, lenShortURL, storage, r.Body)
		if err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write(shortURL)
	}
}

func DecodeURL(storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var shortURL string
		id := chi.URLParam(r, "id")
		if id != "" {
			shortURL = id
		} else {
			// обратная совместимость
			shortURL = r.URL.Path[len("/"):]
		}

		url, err := usecase.DecodeURL(r.Context(), storage, shortURL)
		if err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(url))
	}
}

func EncondeURLJson(baseURL string, lenShortURL int, storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		responce, err := usecase.EncondeURLJson(r.Context(), baseURL, lenShortURL, storage, r.Body)
		if err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responce)
	}
}

func PingDatabase(dsn string, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := usecase.PingDatabase(r.Context(), dsn, timeout); err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
