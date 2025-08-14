// Package config предоставляет функциональность для работы с конфигурацией приложения.
// Включает загрузку конфигурации из файлов, переменных окружения и флагов командной строки.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// Config - модель конфигурации приложения
type Config struct {
	// ListenAddr - адрес сервера
	ListenAddr string `env:"SERVER_ADDRESS" json:"server_address"`
	// BaseURL - базовый URL для формирования коротких ссылок
	BaseURL string `env:"BASE_URL" json:"base_url"`
	// ShortURLLen - длинна сгенерированных коротких ссылок
	ShortURLLen int `env:"MAX_URL_LEN" json:"max_url_len"`
	// LogLevel - уровени логирования
	LogLevel string `env:"LOG_LEVEL" json:"log_level"`
	// FileStoragePath - путь к файловому хранилищу
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	// DatabaseDSN - строка подключения к БД
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// DatabaseTimeout - таймаут запросов к БД
	DatabaseTimeout time.Duration `env:"DATABASE_TIMEOUT"  json:"database_timeout"`
	// JWTSecret - секрет для JWT токена
	JWTSecret string `env:"JWT_SECRET" json:"jwt_secret"`
	// DebugEnable - признак включения отладочного режима (профилирование)
	DebugEnable bool `json:"enable_debug"`
	// HTTPSEnabled - признак включения https
	HTTPSEnabled bool `env:"ENABLE_HTTPS" json:"enable_https"`
	// ConfigFilePath - путь к файлу конфигурации
	ConfigFilePath string `env:"CONFIG" json:"-"`
	// TrustedSubnet - доверенная подсеть
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
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
	DefaultDebugEnabled    = false
	DefaultHTTPSEnabled    = false
	DefaultConfigFilePath  = ""
	DefaultTrustedSubnet   = ""
)

func (cfg *Config) parseFromEnv() {
	if err := env.Parse(cfg); err != nil {
		panic(fmt.Sprintf("Failed to parse enviroment var: %s", err.Error()))
	}
}
func (cfg *Config) parseFromFlags() {

	pflag.StringVarP(&cfg.ListenAddr, "server", "a", DefaultListenServer, "Server listen address in a form host:port.")
	pflag.StringVarP(&cfg.BaseURL, "base_url", "b", DefaultBaseURL, "Server base URL.")
	pflag.IntVar(&cfg.ShortURLLen, "url_len", DefaultShortURLlen, "Short URL length.")
	pflag.StringVar(&cfg.LogLevel, "log_level", DefaultLogLevel, "Log level.")
	pflag.StringVarP(&cfg.FileStoragePath, "file_storage_path", "f", filepath.Join(os.TempDir(), DefaultCacheFileName), "Path to cache file.")
	pflag.StringVarP(&cfg.DatabaseDSN, "db_dsn", "d", DefaultDatabaseDSN, "Database DSN")
	pflag.DurationVar(&cfg.DatabaseTimeout, "db_timeout", DefaultDatabaseTimeout, "Database timeout connection, seconds.")
	pflag.StringVar(&cfg.JWTSecret, "jwt_secret", DefaultJWTSecret, "Secret to JWT")
	pflag.BoolVar(&cfg.DebugEnable, "debug", DefaultDebugEnabled, "Debug mode")
	pflag.BoolVarP(&cfg.HTTPSEnabled, "https", "s", DefaultHTTPSEnabled, "Enable https")
	pflag.StringVarP(&cfg.ConfigFilePath, "config", "c", DefaultConfigFilePath, "Path to config file.")
	pflag.StringVarP(&cfg.TrustedSubnet, "trusted_subnet", "t", DefaultTrustedSubnet, "Trusted subnet")

	pflag.Parse()
}

func (cfg *Config) parseFromFile() {
	if len(cfg.ConfigFilePath) == 0 {
		return
	}
	buf, err := os.ReadFile(cfg.ConfigFilePath)
	if err != nil {
		panic(fmt.Sprintf("can't load config file: %s ", errors.Cause(err).Error()))
	}
	tmp := NewDefaultConfig()
	if err = json.Unmarshal(buf, tmp); err != nil {
		panic(fmt.Sprintf("can't parse config file: %s ", errors.Cause(err).Error()))
	}
	// Определение адреса сервера
	if cfg.ListenAddr == DefaultListenServer {
		cfg.ListenAddr = tmp.ListenAddr
	}
	// Определение базового URL
	if cfg.BaseURL == DefaultBaseURL {
		cfg.BaseURL = tmp.BaseURL
	}
	// Определение максимальной длинны сокращаемых URLs
	if cfg.ShortURLLen == DefaultShortURLlen {
		cfg.ShortURLLen = tmp.ShortURLLen
	}
	// Определение уровня логирования
	if cfg.LogLevel == DefaultLogLevel {
		cfg.LogLevel = tmp.LogLevel
	}
	// Определение пути к файлу с текстовым кэшем
	if cfg.FileStoragePath == DefaultCacheFileName {
		cfg.FileStoragePath = tmp.FileStoragePath
	}
	// Определение DSN для подключения к БД
	if cfg.DatabaseDSN == DefaultDatabaseDSN {
		cfg.DatabaseDSN = tmp.DatabaseDSN
	}
	// Определение таймаута работы с БД
	if cfg.DatabaseTimeout == DefaultDatabaseTimeout {
		cfg.DatabaseTimeout = tmp.DatabaseTimeout
	}
	// Определение секрета JWT токена
	if cfg.JWTSecret == DefaultJWTSecret {
		cfg.JWTSecret = tmp.JWTSecret
	}
	// Определение признака работы в debug (profiler)
	if !cfg.DebugEnable {
		cfg.DebugEnable = tmp.DebugEnable
	}
	// Определение признака работы по HTTPS
	if !cfg.HTTPSEnabled {
		cfg.HTTPSEnabled = tmp.HTTPSEnabled
	}
	// Определение доверенной подсети
	if cfg.TrustedSubnet == DefaultTrustedSubnet {
		cfg.TrustedSubnet = tmp.TrustedSubnet
	}
}

// NewConfig - метод формирования конфигурации приложения. Используются переменные окружения и флаги запуска приложения.
func NewConfig() *Config {

	// инициализируемся по дефолту
	config := NewDefaultConfig()

	// Устанавливаем из флагов
	config.parseFromFlags()

	// Перезаписываем параметры, которые есть в окружении
	config.parseFromEnv()

	// Устанавливаем файла те параметры, которые не установлены ранее
	config.parseFromFile()

	return config
}

// NewDefaultConfig - метод формирования конфигурации по-умолчанию
func NewDefaultConfig() *Config {
	return &Config{
		ListenAddr:      DefaultListenServer,
		BaseURL:         DefaultBaseURL,
		ShortURLLen:     DefaultShortURLlen,
		LogLevel:        DefaultLogLevel,
		FileStoragePath: filepath.Join(os.TempDir(), DefaultCacheFileName),
		DatabaseDSN:     DefaultDatabaseDSN,
		JWTSecret:       DefaultJWTSecret,
		DebugEnable:     DefaultDebugEnabled,
		HTTPSEnabled:    DefaultHTTPSEnabled,
		ConfigFilePath:  DefaultConfigFilePath,
		TrustedSubnet:   DefaultTrustedSubnet,
	}
}
