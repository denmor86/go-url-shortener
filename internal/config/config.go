package config

import (
	"errors"
	"flag"
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

	listenAddr := new(NetAddress)
	flag.Var(listenAddr, "a", "Server listen address in a form host:port.")
	baseURL := flag.String("b", DefaultBaseURL, "Server base URL.")
	ShortURLLen := flag.Int("l", DefaultShortURLlen, "Short URL length.")

	flag.Parse()

	if listenAddr.Host == "" {
		listenAddr = &NetAddress{DefaultListenHost, DefaultListenPort}
	}
	if *baseURL == "" {
		*baseURL = DefaultBaseURL
	}
	if *ShortURLLen > DefaultShortURLlen {
		*ShortURLLen = DefaultShortURLlen
	}

	return &Config{
		ListenAddr:  *listenAddr,
		BaseURL:     *baseURL,
		ShortURLLen: *ShortURLLen,
	}
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr:  NetAddress{DefaultListenHost, DefaultListenPort},
		BaseURL:     DefaultBaseURL,
		ShortURLLen: DefaultShortURLlen,
	}
}
