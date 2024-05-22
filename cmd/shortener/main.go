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

	dbDSNFlag := flag.String("d", "", "Строка подключения к базе данных")
	address := flag.String("a", "", "адрес для запуска HTTP-сервера")
	baseURL := flag.String("b", "", "базовы URL для сокращенных URL")
	fileStoragePath := flag.String("f", "", "путь к файлу для хранения данных")
	flag.Parse()

	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		dbDSN = *dbDSNFlag
	}

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

	cfg.Address = *address
	cfg.BaseURL = *baseURL
	cfg.FileStoragePath = *fileStoragePath
	cfg.DatabaseDSN = dbDSN

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if dbDSN != "" {
		pool, err := pgxpool.Connect(ctx, dbDSN)
		if err != nil {
			logger.Fatalf("Unable to connect to database: %v", err)
		}
		defer pool.Close()
		err = db.InitDB(pool)
		if err != nil {
			logger.Fatalf("Failed to initialize database: %v", err)
		}

		server := server.NewServer(cfg, pool)
		logger.Infof("Запуск сервера на адресе %s", cfg.Address)

		if err := server.Run(); err != nil {
			logger.Fatalf("Ошибка запуска сервера: %v", err)
		}
	} else {
		logger.Error("DATABASE_DSN environment variable and -d flag are not set, using internal storage")
		server := server.NewServer(cfg, nil)
		logger.Infof("Запуск сервера на адресе %s", cfg.Address)

		if err := server.Run(); err != nil {
			logger.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}
}
