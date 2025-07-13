package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"

	"github.com/denmor86/go-url-shortener.git/internal/usecase"
)

func EncondeURL(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userID := r.Context().Value(usecase.UserIDContextKey); userID != nil {
			shortURL, err := u.EncondeURL(r.Context(), r.Body, userID.(string))

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

func EncondeURLJson(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userID := r.Context().Value(usecase.UserIDContextKey); userID != nil {
			responce, err := u.EncondeURLJson(r.Context(), r.Body, userID.(string))

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

func EncondeURLJsonBatch(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userID := r.Context().Value(usecase.UserIDContextKey); userID != nil {
			responce, err := u.EncondeURLJsonBatch(r.Context(), r.Body, userID.(string))
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

func PingStorage(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := u.PingStorage(r.Context()); err != nil {
			http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

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

func DeleteURLS(u *usecase.Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if userID := r.Context().Value(usecase.UserIDContextKey); userID != nil {
			err := u.DeleteURLS(r.Context(), r.Body, userID.(string))
			if err != nil {
				http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
		}
	}
}
