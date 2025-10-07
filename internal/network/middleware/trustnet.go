// Package middleware предоставляет впомогательные middleware методы для поддержки сетевого взаимодействия
package middleware

import (
	"net"
	"net/http"
)

// Внутренние константы middelware
const (
	// tokenIpHeaderKey имя ключа для IP адреса в заголовке запроса
	tokenIPHeaderKey = "X-Real-IP"
)

// Authorization - модель middelware для авторизации пользователя
type TrustNet struct {
	subnet *net.IPNet // секрет для JWT
}

// NewAuthorization - метод формирования объекта middelware для проверки доверенной подсети
func NewTrustNet(subnet *net.IPNet) *TrustNet {
	return &TrustNet{subnet: subnet}
}

// TrustGuard — middleware-проверка доверенной подсети для входящих HTTP-запросов.
func (guard *TrustNet) TrustGuard(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realIP := r.Header.Get(tokenIPHeaderKey)
		if len(realIP) == 0 {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		ip := net.ParseIP(realIP)
		if ip == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if ok := guard.subnet.Contains(ip); ok {
			h.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})
}
