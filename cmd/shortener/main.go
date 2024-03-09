package main

import (
	"fiber-apis/internal/models"
	"fiber-apis/internal/server"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func main() {

	logger := logrus.New()

	// Устанавливаем формат вывода
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Устанавливаем уровень логирования
	logger.SetLevel(logrus.InfoLevel)

	// Получаем строку подключения к БД из переменной окружения
	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		log.Fatal("DATABASE_DSN environment variable is not set")
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

	config := models.Config{
		Address:         *address,
		BaseURL:         *baseURL,
		FileStoragePath: *fileStoragePath,
	}

	server := server.NewServer(config)

	// Запускаем сервер
	err := server.Run()
	if err != nil {
		logger.Fatalf("Error running server: %v", err)
	}
}
