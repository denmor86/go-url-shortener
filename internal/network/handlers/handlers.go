package handlers

import (
	"net/http"

	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

func EncondeURL(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		shortURL, err := u.EncondeURL(r.Context(), r.Body)
		if err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write(shortURL)
	}
}

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
		if err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(url))
	}
}

func EncondeURLJson(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		responce, err := u.EncondeURLJson(r.Context(), r.Body)
		if err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responce)
	}
}

func EncondeURLJsonBatch(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		responce, err := u.EncondeURLJsonBatch(r.Context(), r.Body)
		if err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responce)
	}
}

func PingStorage(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := u.PingStorage(r.Context()); err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
