package main

import (
	"context"
	"fiber-apis/internal/models"
	"fiber-apis/internal/server"
	"fiber-apis/internal/storage"
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	var cfg models.Config

	if err := env.Parse(&cfg); err != nil {
		logrus.Errorf("Ошибка при парсинге переменных окружения: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	dbDSN := flag.String("d", "", "Строка подключения к базе данных")
	flag.Parse()

	var storable models.Storable

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if *dbDSN != "" {
		pool, err := pgxpool.Connect(ctx, *dbDSN)
		if err != nil {
			logger.Errorf("Unable to connect to database: %v", err)
			storable = storage.NewInternalStorage()
		} else {
			defer pool.Close()
			storable = storage.NewDatabaseStorage(pool)
		}
	} else {
		logger.Error("DATABASE_URL environment variable is not set, using internal storage")
		storable = storage.NewInternalStorage()
	}

	address := flag.String("a", "", "адрес для запуска HTTP-сервера")
	baseURL := flag.String("b", "", "базовый URL для сокращенных URL")
	fileStoragePath := flag.String("f", "", "путь к файлу для хранения данных")
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

	server := server.NewServer(config, storable)

	logger.Infof("Запуск сервера на адресе %s", cfg.Address)

	if err := server.Run(); err != nil {
		logger.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
