package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/spf13/pflag"
)

type Config struct {
	ListenAddr  string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	ShortURLLen int    `env:"MAX_URL_LEN" envDefault:"8"`
}

const (
	DefaultListenServer = "localhost:8080"
	DefaultBaseURL      = "http://localhost:8080"
	DefaultShortURLlen  = 8
)

func NewConfig() *Config {

	pflag.StringP("server", "a", DefaultListenServer, "Server listen address in a form host:port.")
	pflag.StringP("base_url", "b", DefaultBaseURL, "Server base URL.")
	pflag.IntP("url_len", "l", DefaultShortURLlen, "Short URL length.")
	pflag.Parse()

	var config Config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("Failed to parse enviroment var: %v", err)
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
		if url_len, err := pflag.CommandLine.GetInt("url_len"); err == nil {
			config.ShortURLLen = url_len
		}
	}

	return &config
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr:  DefaultListenServer,
		BaseURL:     DefaultBaseURL,
		ShortURLLen: DefaultShortURLlen,
	}
}
