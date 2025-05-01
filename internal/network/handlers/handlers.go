package handlers

import (
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func EncondeURL(baseURL string, lenShortURL int, storage usecase.IBaseStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		shortURL, err := usecase.EncondeURL(baseURL, lenShortURL, storage, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write(shortURL)
	}
}

func DecodeURL(storage usecase.IBaseStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var shortURL string
		id := chi.URLParam(r, "id")
		if id != "" {
			shortURL = id
		} else {
			// обратная совместимость
			shortURL = r.URL.Path[len("/"):]
		}

		url, err := usecase.DecodeURL(storage, shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(url))
	}
}

func EncondeURLJson(baseURL string, lenShortURL int, storage usecase.IBaseStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		responce, err := usecase.EncondeURLJson(baseURL, lenShortURL, storage, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responce)
	}
}

func PingStorage(storage usecase.IBaseStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		http.Error(w, "Storage is disconnected", http.StatusInternalServerError)
	}
}
