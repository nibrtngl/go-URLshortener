package main

import (
	"fiber-apis/internal/server"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePath == "" {
		fileStoragePathFlag := flag.String("f", "/tmp/KFSBM", "Path to the file for storing data")
		fileStoragePath = *fileStoragePathFlag // Dereference the pointer to get the string value
	}

	if *address == "" {
		*address = "localhost:8080"
	}
	if *baseURL == "" {
		*baseURL = "http://localhost:8080"
	}

	// Create the directory for file storage if it doesn't exist
	err := os.MkdirAll(filepath.Dir(fileStoragePath), os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory: %v", err)
		return
	}

	// Create the file for data storage if it doesn't exist
	_, err = os.OpenFile(fileStoragePath, os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error creating file: %v", err)
		return
	}

	config := server.Config{
		Address:         *address,
		BaseURL:         *baseURL,
		FileStoragePath: fileStoragePath,
	}

	server := server.NewServer(config)

	// Run the server
	err = server.Run()
	if err != nil {
		fmt.Printf("Error running server: %v", err)
	}

	// Read the file with saved URLs
	data, err := ioutil.ReadFile(fileStoragePath)
	if err != nil {
		fmt.Printf("Error reading file: %v", err)
		return
	}

	fmt.Println("Saved URLs:")
	fmt.Println(string(data))
}
