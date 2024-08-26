package main

import (
	"context"
	"fiber-apis/config"
	"fiber-apis/internal/db"
	"fiber-apis/internal/models"
	"fiber-apis/internal/server"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/gorilla/securecookie"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/pprof"
	"time"
)

func main() {
	var configFilePath string

	var buildVersion string
	var buildDate string
	var buildCommit string

	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	var cfg models.Config

	flag.StringVar(&configFilePath, "c", "", "Path to JSON config file")
	dbDSNFlag := flag.String("d", "", "Database connection string")
	address := flag.String("a", "", "HTTP server address")
	baseURL := flag.String("b", "", "Base URL for shortened URLs")
	fileStoragePath := flag.String("f", "", "Path to file storage")
	enableHTTPS := flag.Bool("s", false, "Enable HTTPS")
	certFile := flag.String("cert", "", "Path to SSL certificate")
	keyFile := flag.String("key", "", "Path to SSL key")
	flag.Parse()

	// Загрузка конфигурации из файла, если указан
	if configFilePath != "" {
		fileCfg, err := config.LoadConfigFromFile(configFilePath)
		if err != nil {
			logrus.Fatalf("Error loading config file: %v", err)
		}
		cfg = *fileCfg
	}

	// Перегружаем переменные окружения
	if err := env.Parse(&cfg); err != nil {
		logrus.Errorf("Ошибка при парсинге переменных окружения: %v", err)
	}

	// Перегружаем значения из флагов командной строки
	if *address != "" {
		cfg.Address = *address
	}
	if *baseURL != "" {
		cfg.BaseURL = *baseURL
	}
	if *fileStoragePath != "" {
		cfg.FileStoragePath = *fileStoragePath
	}
	if *dbDSNFlag != "" {
		cfg.DatabaseDSN = *dbDSNFlag
	}
	if *enableHTTPS {
		cfg.EnableHTTPS = true
	}

	// Логгер
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	// Профайлинг через pprof
	go func() {
		pprofMux := http.NewServeMux()
		pprofMux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		pprofMux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		pprofMux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		pprofMux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		pprofMux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		http.ListenAndServe("localhost:6060", pprofMux)
	}()

	// Инициализация базы данных
	var pool *pgxpool.Pool
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if cfg.DatabaseDSN != "" {
		pool, err = pgxpool.Connect(ctx, cfg.DatabaseDSN)
		if err != nil {
			logger.Fatalf("Unable to connect to database: %v", err)
		}
		defer pool.Close()
		err = db.InitDB(pool)
		if err != nil {
			logger.Fatalf("Failed to initialize database: %v", err)
		}
	} else {
		logger.Info("DATABASE_DSN is not set, using internal storage")
	}

	// Инициализация сервера
	secureCookie := securecookie.New([]byte("very-secret"), []byte("a-lot-secret"))
	srv := server.NewServer(cfg, pool, secureCookie)

	// Запуск сервера
	logger.Infof("Запуск сервера на адресе %s", cfg.Address)
	if cfg.EnableHTTPS {
		if *certFile == "" || *keyFile == "" {
			logger.Fatal("Both cert and key file must be specified for HTTPS")
		}
		if err := srv.RunTLS(*certFile, *keyFile); err != nil {
			logger.Fatalf("Ошибка запуска HTTPS сервера: %v", err)
		}
	} else {
		if err := srv.Run(); err != nil {
			logger.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}
}
