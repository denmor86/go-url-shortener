package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/denmor86/go-url-shortener.git/internal/models"
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

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(helpers.MakeURL(baseURL, shortURL)))
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

func EncondeURLJsonHandler(baseURL string, lenShortURL int, storage storage.IStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var buf bytes.Buffer
		// читаем тело запроса
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var request models.Request
		if err = json.Unmarshal(buf.Bytes(), &request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		url := request.URL

		if len(url) == 0 {
			http.Error(w, "URL is empty", http.StatusBadRequest)
			return
		}

		shortURL := helpers.MakeShortURL(url, lenShortURL)
		storage.Add(url, shortURL)

		var responce models.Response
		responce.Result = helpers.MakeURL(baseURL, shortURL)
		resp, err := json.Marshal(responce)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	}
}
