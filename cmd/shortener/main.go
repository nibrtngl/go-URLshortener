package main

import (
	"context"
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
		logrus.Errorf("Ошибка при парсинге переменных окружения: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	address := flag.String("a", "", "адрес для запуска HTTP-сервера")
	baseURL := flag.String("b", "", "базовый URL для сокращенных URL")
	dbDSN := flag.String("d", "", "строка подключения к базе данных")
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
	if *dbDSN == "" {
		*dbDSN = os.Getenv("DATABASE_DSN")
	}
	config := models.Config{
		Address:         *address,
		BaseURL:         *baseURL,
		FileStoragePath: *fileStoragePath,
		DatabaseDSN:     *dbDSN,
	}

	var storable models.Storable

	if *dbDSN != "" {
		poolConfig, err := pgxpool.ParseConfig(*dbDSN)
		if err != nil {
			log.Fatalf("Ошибка при получении параметров подключения к базе данных: %v", err)
		}

		pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
		if err != nil {
			logrus.Errorf("Ошибка подключения к базе данных: %v", err)
		}
		defer pool.Close()

	}

	server := server.NewServer(config, storable)

	logger.Infof("Запуск сервера на адресе %s", cfg.Address)

	if err := server.Run(); err != nil {
		logrus.Errorf("Ошибка запуска сервера: %v", err)
	}
}
