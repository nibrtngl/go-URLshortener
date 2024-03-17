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
	"log"
	"os"
)

func main() {
	var cfg models.Config

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка при парсинге переменных окружения: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	dbDSN := flag.String("d", "", "строка подключения к базе данных")
	flag.Parse()

	if *dbDSN == "" {
		*dbDSN = os.Getenv("DATABASE_DSN")
	}

	var storable models.Storable

	if *dbDSN != "" {
		poolConfig, err := pgxpool.ParseConfig(*dbDSN)
		if err != nil {
			log.Fatalf("Ошибка при получении параметров подключения к базе данных: %v", err)
		}

		pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
		if err != nil {
			log.Fatalf("Ошибка подключения к базе данных: %v", err)
		}
		defer pool.Close()

		storable = storage.NewDatabaseStorage(pool)
	} else {
		storable = storage.NewInMemoryStorage()
	}

	server := server.NewServer(cfg, storable)

	logger.Infof("Запуск сервера на адресе %s", cfg.Address)

	if err := server.Run(); err != nil {
		logger.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
