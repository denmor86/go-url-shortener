package config

import (
	"github.com/caarlos0/env"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
	"github.com/spf13/pflag"
)

type Config struct {
	ListenAddr  string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	ShortURLLen int    `env:"MAX_URL_LEN" envDefault:"8"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}

const (
	DefaultListenServer = "localhost:8080"
	DefaultBaseURL      = "http://" + DefaultListenServer
	DefaultShortURLlen  = 8
	DefaultLogLevel     = "info"
)

func NewConfig() *Config {

	pflag.StringP("server", "a", DefaultListenServer, "Server listen address in a form host:port.")
	pflag.StringP("base_url", "b", DefaultBaseURL, "Server base URL.")
	pflag.IntP("url_len", "s", DefaultShortURLlen, "Short URL length.")
	pflag.StringP("log_level", "l", DefaultLogLevel, "Log level.")
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
		if baseUrl, err := pflag.CommandLine.GetString("base_url"); err == nil {
			config.BaseURL = baseUrl
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
	return &config
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr:  DefaultListenServer,
		BaseURL:     DefaultBaseURL,
		ShortURLLen: DefaultShortURLlen,
		LogLevel:    DefaultLogLevel,
	}
}
