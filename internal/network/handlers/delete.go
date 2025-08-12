// Package handlers предоставляет описание обработчиков сетевых сообщений
package handlers

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/denmor86/go-url-shortener/internal/usecase"
)

// DeleteURLS - метод-обработчик формирования запроса на удаление данных о сокращенных URL пользователя
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
			return
		}
		http.Error(w, "Undefined user", http.StatusBadRequest)
	}
}
