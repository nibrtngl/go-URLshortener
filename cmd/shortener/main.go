// main.go
package main

import (
	"context"
	"fiber-apis/internal/db"
	"fiber-apis/internal/models"
	"fiber-apis/internal/server"
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func main() {
	var cfg models.Config

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error Parsing Environment variables: %v", err)
	}

	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logger.SetLevel(logrus.InfoLevel)

	// Получение параметров подключения к базе данных
	dbHost, dbPort, dbUser, dbPassword, dbName := db.GetDBConnectionParams()

	// Формирование строки подключения к базе данных
	dbDSN := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName

	// Подключение к базе данных
	poolConfig, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		log.Fatalf("Error parsing database DSN: %v", err)
	}

	// Используем IPv4-адрес 127.0.0.1
	poolConfig.ConnConfig.Host = "127.0.0.1"

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer pool.Close()

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

	server := server.NewServer(config, pool)

	// Запускаем сервер
	err = server.Run()
	if err != nil {
		logger.Fatalf("Error running server: %v", err)
	}
}
