package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int16
}

type Config struct {
	ListenAddr  NetAddress
	BaseURL     string
	LenShortURL int
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
	a.Port = int16(port)
	return nil
}

func NewConfig() *Config {

	listenAddr := new(NetAddress)
	flag.Var(listenAddr, "a", "Server listen address in a form host:port.")

	baseURL := flag.String("b", "http://localhost:8080/", "Server base URL.")
	if *baseURL == "" {
		*baseURL = "http://localhost:8080"
	}

	lenShortURL := flag.Int("l", 8, "Len short URL.")
	if *lenShortURL > 8 {
		*lenShortURL = 8
	}

	flag.Parse()

	return &Config{
		ListenAddr:  *listenAddr,
		BaseURL:     *baseURL,
		LenShortURL: *lenShortURL,
	}
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr: NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		BaseURL:     "http://localhost:8080/",
		LenShortURL: 8,
	}
}
