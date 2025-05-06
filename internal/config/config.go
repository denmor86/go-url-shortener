package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env"
	"github.com/spf13/pflag"
)

type Config struct {
	ListenAddr      string        `env:"SERVER_ADDRESS"`
	BaseURL         string        `env:"BASE_URL"`
	ShortURLLen     int           `env:"MAX_URL_LEN" envDefault:"8"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string        `env:"DATABASE_DSN"`
	DatabaseTimeout time.Duration `env:"DATABASE_TIMEOUT"`
}

const (
	DefaultListenServer    = "localhost:8080"
	DefaultBaseURL         = "http://" + DefaultListenServer
	DefaultShortURLlen     = 8
	DefaultLogLevel        = "info"
	DefaultCacheFileName   = "shortener_cache.txt"
	DefaultDatabaseDSN     = ""
	DefaultDatabaseTimeout = time.Duration(5)
)

func NewConfig() Config {

	pflag.StringP("server", "a", DefaultListenServer, "Server listen address in a form host:port.")
	pflag.StringP("base_url", "b", DefaultBaseURL, "Server base URL.")
	pflag.IntP("url_len", "s", DefaultShortURLlen, "Short URL length.")
	pflag.StringP("log_level", "l", DefaultLogLevel, "Log level.")
	pflag.StringP("file_storage_path", "f", filepath.Join(os.TempDir(), DefaultCacheFileName), "Path to cache file.")
	pflag.StringP("db_dsn", "d", DefaultDatabaseDSN, "Database DSN")
	pflag.IntP("db_timeout", "t", int(DefaultDatabaseTimeout.Abs()), "Database timeout connection, seconds.")
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
	return config
}

func DefaultConfig() Config {
	return Config{
		ListenAddr:      DefaultListenServer,
		BaseURL:         DefaultBaseURL,
		ShortURLLen:     DefaultShortURLlen,
		LogLevel:        DefaultLogLevel,
		FileStoragePath: "",
		DatabaseDSN:     DefaultDatabaseDSN,
	}
}
