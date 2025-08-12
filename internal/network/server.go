package network

import (
	"crypto/tls"
	"net/http"

	"github.com/denmor86/go-url-shortener/internal/helpers"
	"github.com/denmor86/go-url-shortener/internal/logger"
	"github.com/denmor86/go-url-shortener/internal/network/router"
	"github.com/denmor86/go-url-shortener/internal/usecase"
)

// StartServer - метод запускает сервер в зависимости от настроек http/https
func StartServer(server *http.Server, https bool) error {
	if https {
		return server.ListenAndServeTLS("", "")
	}
	return server.ListenAndServe()
}

// NewServer - метод создаёт новый сервер
func NewServer(listenAddr string, use *usecase.Usecase) *http.Server {
	// Генерируем самоподписанный сертификат
	cert, key, err := helpers.GenerateSelfSignedCert()
	if err != nil {
		logger.Error("error generate certificate", err.Error())
	}

	// Создаем TLS конфигурацию
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert},
				PrivateKey:  key,
			},
		},
	}

	return &http.Server{
		Addr:      listenAddr,
		Handler:   router.HandleRouter(use),
		TLSConfig: tlsConfig,
	}
}
