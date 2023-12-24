package main

import (
	"fiber-apis/internal/server"
	"flag"
	"fmt"
	"os"
)

func main() {
	address := flag.String("a", "", "address to run the HTTP server")
	baseURL := flag.String("b", "", "base URL for the shortened URL")
	flag.Parse()

	// Read environment variables
	if envAddress := os.Getenv("SERVER_ADDRESS"); envAddress != "" {
		*address = envAddress
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		*baseURL = envBaseURL
	}

	// Set default values if not provided
	if *address == "" {
		*address = "localhost:8080"
	}
	if *baseURL == "" {
		*baseURL = "http://localhost:8080"
	}

	config := server.Config{
		Address: *address,
		BaseURL: *baseURL,
	}

	server := server.NewServer(config)

	// Run the server
	err := server.Run()
	if err != nil {
		fmt.Printf("Error running server: %v", err)
	}
}
