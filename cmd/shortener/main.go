package main

import (
	"fiber-apis/internal/server"
	"flag"
	"fmt"
)

func main() {

	address := flag.String("a", "localhost:8080", "address to run the HTTP server")
	baseURL := flag.String("b", "http://localhost:8080", "base URL for the shortened URL")
	flag.Parse()

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
