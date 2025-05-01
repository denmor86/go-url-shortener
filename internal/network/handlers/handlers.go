package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/helpers"
	"github.com/go-chi/chi/v5"
)

type IBaseStorage interface {
	Add(string, string) error
	Get(string) (string, error)
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func EncondeURL(baseURL string, lenShortURL int, storage IBaseStorage) http.HandlerFunc {
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

func DecodeURL(storage IBaseStorage) http.HandlerFunc {
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

func EncondeURLJson(baseURL string, lenShortURL int, storage IBaseStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var buf bytes.Buffer
		// читаем тело запроса
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var request Request
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

		var responce Response
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

func PingStorage(storage IBaseStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		http.Error(w, "Storage is disconnected", http.StatusInternalServerError)
	}
}
