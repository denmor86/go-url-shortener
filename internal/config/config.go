// Package config предоставляет функциональность для работы с конфигурацией приложения.
// Включает загрузку конфигурации из файлов, переменных окружения и флагов командной строки.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env"
	"github.com/spf13/pflag"
)

// Config - модель конфигурации приложения
type Config struct {
	// ListenAddr - адрес сервера
	ListenAddr string `env:"SERVER_ADDRESS"`
	// BaseURL - базовый URL для формирования коротких ссылок
	BaseURL string `env:"BASE_URL"`
	// ShortURLLen - длинна сгенерированных коротких ссылок
	ShortURLLen int `env:"MAX_URL_LEN" envDefault:"8"`
	// LogLevel - уровени логирования
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	// FileStoragePath - путь к файловому хранилищу
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// DatabaseDSN - строка подключения к БД
	DatabaseDSN string `env:"DATABASE_DSN"`
	// DatabaseTimeout - таймаут запросов к БД
	DatabaseTimeout time.Duration `env:"DATABASE_TIMEOUT"`
	// JWTSecret - секрет для JWT токена
	JWTSecret string `env:"JWT_SECRET"`
	// UseDebug - признак включения отладочного режима (профилирование)
	UseDebug bool
}

// Настройки по-умолчанию
const (
	DefaultListenServer    = "localhost:8080"
	DefaultBaseURL         = "http://" + DefaultListenServer
	DefaultShortURLlen     = 8
	DefaultLogLevel        = "info"
	DefaultCacheFileName   = "shortener_cache.txt"
	DefaultDatabaseDSN     = ""
	DefaultDatabaseTimeout = time.Duration(5)
	DefaultJWTSecret       = "secret"
	DefaultUseDebug        = false
)

// NewConfig - метод формирования конфигурации приложения. Используются переменные окружения и флаги запуска приложения.
func NewConfig() Config {

	pflag.StringP("server", "a", DefaultListenServer, "Server listen address in a form host:port.")
	pflag.StringP("base_url", "b", DefaultBaseURL, "Server base URL.")
	pflag.IntP("url_len", "n", DefaultShortURLlen, "Short URL length.")
	pflag.StringP("log_level", "l", DefaultLogLevel, "Log level.")
	pflag.StringP("file_storage_path", "f", filepath.Join(os.TempDir(), DefaultCacheFileName), "Path to cache file.")
	pflag.StringP("db_dsn", "d", DefaultDatabaseDSN, "Database DSN")
	pflag.IntP("db_timeout", "t", int(DefaultDatabaseTimeout.Abs()), "Database timeout connection, seconds.")
	pflag.StringP("jwt_secret", "s", DefaultJWTSecret, "Secret to JWT")
	pflag.BoolP("debug", "m", DefaultUseDebug, "Debug mode")

	pflag.Parse()

	var config Config
	if err := env.Parse(&config); err != nil {
		panic(fmt.Sprintf("Failed to parse enviroment var: %s", err.Error()))
	}

	if config.ListenAddr == "" {
		if address, err := pflag.CommandLine.GetString("server"); err == nil {
			config.ListenAddr = address
		}
	}
	if config.BaseURL == "" {
		if baseURL, err := pflag.CommandLine.GetString("base_url"); err == nil {
			config.BaseURL = baseURL
		}
	}
	if config.ShortURLLen == DefaultShortURLlen {
		if urlLen, err := pflag.CommandLine.GetInt("url_len"); err == nil {
			config.ShortURLLen = urlLen
		}
	}
	if config.LogLevel == DefaultLogLevel {
		if logLevel, err := pflag.CommandLine.GetString("log_level"); err == nil {
			config.LogLevel = logLevel
		}
	}
	if config.FileStoragePath == "" {
		if filepath, err := pflag.CommandLine.GetString("file_storage_path"); err == nil {
			config.FileStoragePath = filepath
		}
	}
	if config.DatabaseDSN == "" {
		if dsn, err := pflag.CommandLine.GetString("db_dsn"); err == nil {
			config.DatabaseDSN = dsn
		}
	}
	if config.DatabaseTimeout == 0 {
		if timeout, err := pflag.CommandLine.GetInt("db_timeout"); err == nil {
			config.DatabaseTimeout = time.Duration(timeout) * time.Second
		}
	}
	if config.JWTSecret == "" {
		if secret, err := pflag.CommandLine.GetString("jwt_secret"); err == nil {
			config.JWTSecret = secret
		}
	}

	if debug, err := pflag.CommandLine.GetBool("debug"); err == nil {
		config.UseDebug = debug
	}

	return config
}

// DefaultConfig - метод формирования конфигурации по-умолчанию
func DefaultConfig() Config {
	return Config{
		ListenAddr:      DefaultListenServer,
		BaseURL:         DefaultBaseURL,
		ShortURLLen:     DefaultShortURLlen,
		LogLevel:        DefaultLogLevel,
		FileStoragePath: "",
		DatabaseDSN:     DefaultDatabaseDSN,
		JWTSecret:       DefaultJWTSecret,
		UseDebug:        true,
	}
}
