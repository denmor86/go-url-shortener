package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
)

func EncondeUrlHandler(storage storage.IStorage) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method != http.MethodPost {
				http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
				return
			}

			url, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			baseUrl := string(url)

			if len(baseUrl) == 0 {
				http.Error(w, "URL is empty", http.StatusBadRequest)
				return
			}
			shortUrl := helpers.MakeShortUrl(baseUrl, 8)
			storage.Save(baseUrl, shortUrl)

			w.Header().Set("content-type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			w.Write(fmt.Appendf(nil, "http://%s/%s", r.Host, shortUrl))
		})
}

func DecodeUrlHandler(storage storage.IStorage) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method != http.MethodGet {
				http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
				return
			}
			shortUrl := r.URL.Path[len("/"):]
			if len(shortUrl) == 0 {
				http.Error(w, "URL is empty", http.StatusBadRequest)
				return
			}
			baseUrl, err := storage.Load(shortUrl)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Location", baseUrl)
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte(baseUrl))
		})
}
