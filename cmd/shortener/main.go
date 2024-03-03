package main

import (
	"context"
	"fiber-apis/internal/server"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Получаем строку подключения к БД из переменной окружения
	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		log.Fatal("DATABASE_DSN environment variable is not set")
	}

	// Подключаемся к БД
	db, err := server.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(context.Background())

	// Проверяем соединение с базой данных
	err = server.PingDB()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	address := flag.String("a", "", "address to run the HTTP server")
	baseURL := flag.String("b", "", "base URL for the shortened URL")
	fileStoragePath := flag.String("f", "", "path to the file storage")
	flag.Parse()

	if *fileStoragePath == "" {
		*fileStoragePath = os.Getenv("FILE_STORAGE_PATH")
	}

	if *fileStoragePath == "" {
		*fileStoragePath = "/tmp/short-url-db.json"
	}

	fmt.Println("File storage path:", *fileStoragePath)

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
	err = server.Run()
	if err != nil {
		fmt.Printf("Error running server: %v", err)
	}
}
