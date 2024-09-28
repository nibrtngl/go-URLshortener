package main

import (
	"context"
	"fiber-apis/config"
	"fiber-apis/internal/db"
	"fiber-apis/internal/grpc"
	"fiber-apis/internal/models"
	"fiber-apis/internal/server"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/gorilla/securecookie"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var configFilePath string
	var buildVersion, buildDate, buildCommit string

	// Build information
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	var cfg models.Config

	// Parse flags
	flag.StringVar(&configFilePath, "c", "", "Path to JSON config file")
	dbDSNFlag := flag.String("d", "", "Database connection string")
	address := flag.String("a", "", "HTTP server address")
	baseURL := flag.String("b", "", "Base URL for shortened URLs")
	fileStoragePath := flag.String("f", "", "Path to file storage")
	enableHTTPS := flag.Bool("s", false, "Enable HTTPS")
	certFile := flag.String("cert", "", "Path to SSL certificate")
	keyFile := flag.String("key", "", "Path to SSL key")
	trustedSubnetFlag := flag.String("t", "", "Trusted subnet in CIDR format")
	enableGRPC := flag.Bool("g", false, "Enable gRPC server")          // Новый флаг для gRPC
	grpcPort := flag.String("grpc_port", ":50051", "gRPC server port") // Порт для gRPC-сервера
	flag.Parse()

	// Load config from file if provided
	if *trustedSubnetFlag != "" {
		cfg.TrustedSubnet = *trustedSubnetFlag
	}
	if configFilePath != "" {
		fileCfg, err := config.LoadConfigFromFile(configFilePath)
		if err != nil {
			logrus.Fatalf("Error loading config file: %v", err)
		}
		cfg = *fileCfg
	}

	// Parse environment variables
	if err := env.Parse(&cfg); err != nil {
		logrus.Errorf("Error parsing environment variables: %v", err)
	}

	// Override with command-line flags
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

	// Logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logger.SetLevel(logrus.InfoLevel)

	// pprof for profiling
	go func() {
		pprofMux := http.NewServeMux()
		pprofMux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		http.ListenAndServe("localhost:6060", pprofMux)
	}()

	// Initialize database connection
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

	// Initialize DatabaseStorage
	var storage server.Storable
	if cfg.DatabaseDSN != "" {
		storage = db.NewDatabaseStorage(pool)
	} else {
		// Инициализируйте другой тип хранилища, если необходимо
	}

	// Initialize HTTP server
	secureCookie := securecookie.New([]byte("very-secret"), []byte("a-lot-secret"))
	srv := server.NewServer(cfg, pool, secureCookie)

	// Channel for shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Start HTTP server in a goroutine
	go func() {
		logger.Infof("Starting HTTP server on %s", cfg.Address)
		if cfg.EnableHTTPS {
			if *certFile == "" || *keyFile == "" {
				logger.Fatal("Both cert and key files must be specified for HTTPS")
			}
			if err := srv.RunTLS(*certFile, *keyFile); err != nil {
				logger.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			if err := srv.Run(); err != nil {
				logger.Fatalf("Failed to start HTTP server: %v", err)
			}
		}
	}()

	// Start gRPC server if enabled
	if *enableGRPC {
		go func() {
			lis, err := net.Listen("tcp", *grpcPort)
			if err != nil {
				logger.Fatalf("Failed to listen on gRPC port: %v", err)
			}

			grpcServer, err := grpc.GetGRPCServer(cfg, make(chan []string), storage) // Передаем storage вместо pool
			if err != nil {
				logger.Fatalf("Failed to initialize gRPC server: %v", err)
			}

			logger.Infof("Starting gRPC server on %s", *grpcPort)
			if err := grpcServer.Serve(lis); err != nil {
				logger.Fatalf("Failed to start gRPC server: %v", err)
			}
		}()
	}

	logger.Info("Waiting for shutdown signal...")

	<-stop
	logger.Info("Shutdown signal received, shutting down servers...")

	// Shutdown HTTP server
	if err := srv.App.Shutdown(); err != nil {
		logger.Fatalf("Error during HTTP server shutdown: %v", err)
	}

	// Save data to storage if necessary
	if cfg.FileStoragePath != "" {
		if err := srv.SaveData(cfg.FileStoragePath); err != nil {
			logger.Errorf("Error saving data to file: %v", err)
		}
	}

	logger.Info("Servers stopped gracefully")
}
