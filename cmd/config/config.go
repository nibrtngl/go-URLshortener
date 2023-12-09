package config

import "flag"

type ServerConfig struct {
	Address string
	BaseURL string
}

func ParseFlags() ServerConfig {
	address := flag.String("a", "localhost:8080", "address to run the HTTP server")
	baseURL := flag.String("b", "http://localhost:8080", "base URL for the shortened URL")
	flag.Parse()

	return ServerConfig{
		Address: *address,
		BaseURL: *baseURL,
	}
}
