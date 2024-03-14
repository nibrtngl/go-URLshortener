package main

import (
	"context"
	"fiber-apis/internal/models"
	"fiber-apis/internal/server"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func main() {

	err := os.Setenv("DATABASE_DSN", "host=localhost port=5432 user=postgres dbname=postgres sslmode=disable")
	if err != nil {
		fmt.Println("Ошибка при установке переменной окружения:", err)
		return
	}

	var cfg models.Config

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка при парсинге переменных окружения: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	poolConfig, err := getDBConnectionParams(&cfg)
	if err != nil {
		log.Fatalf("Ошибка при получении параметров подключения к базе данных: %v", err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer pool.Close()

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
	if *dbDSN == "" {
		log.Fatal("Не удалось получить параметры подключения к базе данных")
	}

	config := models.Config{
		Address:         *address,
		BaseURL:         *baseURL,
		FileStoragePath: *fileStoragePath,
		DatabaseDSN:     *dbDSN,
	}

	server := server.NewServer(config, pool)

	logger.Infof("Запуск сервера на адресе %s", config.Address)

	if err := server.Run(); err != nil {
		logger.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func getDBConnectionParams(cfg *models.Config) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	return poolConfig, nil
}
