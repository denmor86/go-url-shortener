package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

func EncondeURLHandler(storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
			return
		}

		url, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		baseURL := string(url)

		if len(baseURL) == 0 {
			http.Error(w, "URL is empty", http.StatusBadRequest)
			return
		}
		shortURL := helpers.MakeShortURL(baseURL, 8)
		storage.Save(baseURL, shortURL)

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write(fmt.Appendf(nil, "http://%s/%s", r.Host, shortURL))
	}
}

func DecodeURLHandler(storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
			return
		}

		var shortURL string
		id := chi.URLParam(r, "id")
		if id != "" {
			shortURL = id
		} else {
			// обратная совместимость
			shortURL = r.URL.Path[len("/"):]
		}

		if shortURL == "" {
			http.Error(w, "URL is empty", http.StatusBadRequest)
			return
		}

		baseURL, err := storage.Load(shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", baseURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(baseURL))
	}
}
