package config

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/env"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
	"github.com/spf13/pflag"
)

type Config struct {
	ListenAddr      string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	ShortURLLen     int    `env:"MAX_URL_LEN" envDefault:"8"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

const (
	DefaultListenServer  = "localhost:8080"
	DefaultBaseURL       = "http://" + DefaultListenServer
	DefaultShortURLlen   = 8
	DefaultLogLevel      = "info"
	DefaultCacheFileName = "shortener_cache.txt"
)

func NewConfig() *Config {

	pflag.StringP("server", "a", DefaultListenServer, "Server listen address in a form host:port.")
	pflag.StringP("base_url", "b", DefaultBaseURL, "Server base URL.")
	pflag.IntP("url_len", "s", DefaultShortURLlen, "Short URL length.")
	pflag.StringP("log_level", "l", DefaultLogLevel, "Log level.")
	pflag.StringP("file_storage_path", "f", filepath.Join(os.TempDir(), DefaultCacheFileName), "Path to cache file.")
	pflag.Parse()

	var config Config
	if err := env.Parse(&config); err != nil {
		logger.Error("Failed to parse enviroment var: ", err)
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
	return &config
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr:      DefaultListenServer,
		BaseURL:         DefaultBaseURL,
		ShortURLLen:     DefaultShortURLlen,
		LogLevel:        DefaultLogLevel,
		FileStoragePath: filepath.Join(os.TempDir(), DefaultCacheFileName),
	}
}
