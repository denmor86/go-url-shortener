package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port uint16
}

type Config struct {
	ListenAddr  NetAddress
	BaseURL     string
	ShortURLLen int
}

func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(int(a.Port))
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = uint16(port)
	return nil
}

const (
	DefaultListenHost  = "localhost"
	DefaultListenPort  = 8080
	DefaultBaseURL     = "http://localhost:8080"
	DefaultShortURLlen = 8
)

func NewConfig() *Config {

	var listenAddr NetAddress
	flag.Var(&listenAddr, "a", "Server listen address in a form host:port.")

	var baseURL string
	flag.StringVar(&baseURL, "b", DefaultBaseURL, "Server base URL.")

	var shortURLLen int
	flag.IntVar(&shortURLLen, "l", DefaultShortURLlen, "Short URL length.")

	flag.Parse()

	if listenAddrEnv := os.Getenv("SERVER_ADDRESS"); listenAddrEnv != "" {
		if err := listenAddr.Set(listenAddrEnv); err != nil {
			listenAddr = NetAddress{DefaultListenHost, DefaultListenPort}
		}
	}

	if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
		baseURL = baseURLEnv
	}

	if listenAddr.Host == "" {
		listenAddr = NetAddress{DefaultListenHost, DefaultListenPort}
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if shortURLLen > DefaultShortURLlen {
		shortURLLen = DefaultShortURLlen
	}

	return &Config{
		ListenAddr:  listenAddr,
		BaseURL:     baseURL,
		ShortURLLen: shortURLLen,
	}
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr:  NetAddress{DefaultListenHost, DefaultListenPort},
		BaseURL:     DefaultBaseURL,
		ShortURLLen: DefaultShortURLlen,
	}
}
