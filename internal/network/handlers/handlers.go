package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

func EncondeURLHandler(baseURL string, lenShortURL int, storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		url := string(data)

		if len(url) == 0 {
			http.Error(w, "URL is empty", http.StatusBadRequest)
			return
		}

		shortURL := helpers.MakeShortURL(url, lenShortURL)
		storage.Add(url, shortURL)

		makeURL := func(baseURL, shortURL string) string {
			var fullURL string
			fullURL = baseURL
			if !strings.HasSuffix(fullURL, "/") {
				fullURL += "/"
			}
			fullURL += shortURL
			return fullURL
		}

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(makeURL(baseURL, shortURL)))
	}
}

func DecodeURLHandler(storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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

		url, err := storage.Get(shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(url))
	}
}
