package server

import (
	"context"
	"errors"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"os"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{
		pool: pool,
	}
}

func (s *PostgresStorage) GetURL(short_url string) (string, error) {
	var original_url string
	query := "SELECT original_url FROM urls WHERE short_url=$1"
	err := s.pool.QueryRow(context.Background(), query, short_url).Scan(&original_url)
	if err != nil {
		return "", err
	}
	return original_url, nil
}

func (s *PostgresStorage) SetURL(short_url, original_url string) {
	query := "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)"
	_, err := s.pool.Exec(context.Background(), query, short_url, original_url)
	if err != nil {
		logrus.Errorf("Failed to insert URL into database: %v", err)
	}
}

func (s *PostgresStorage) GetAllKeys() ([]string, error) {
	query := "SELECT short_url FROM urls"
	rows, err := s.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var short_urls []string
	for rows.Next() {
		var short_url string
		if err := rows.Scan(&short_url); err != nil {
			return nil, err
		}
		short_urls = append(short_urls, short_url)
	}
	return short_urls, nil
}

func (s *PostgresStorage) Ping() error {
	return s.pool.Ping(context.Background())
}

func (s *PostgresStorage) CreateTable() error {
	_, err := s.pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS urls (id VARCHAR(255) PRIMARY KEY, url TEXT NOT NULL)")
	return err
}

type InternalStorage struct {
	urls map[string]string
}

func NewInternalStorage() *InternalStorage {
	return &InternalStorage{
		urls: make(map[string]string),
	}
}

func (s *InternalStorage) GetURL(id string) (string, error) {

	originalURL, ok := s.urls[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (s *InternalStorage) SetURL(id, url string) {
	s.urls[id] = url
}

func (s *InternalStorage) GetAllKeys() ([]string, error) {
	keys := make([]string, 0, len(s.urls))
	for k := range s.urls {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *InternalStorage) Ping() error {
	return nil
}

type DatabaseStorage struct {
	pool *pgxpool.Pool
}

func NewDatabaseStorage(pool *pgxpool.Pool) *DatabaseStorage {
	return &DatabaseStorage{
		pool: pool,
	}
}

func (s *DatabaseStorage) GetURL(id string) (string, error) {
	return "", nil
}

func (s *DatabaseStorage) SetURL(id, url string) {

}

func (s *DatabaseStorage) GetAllKeys() ([]string, error) {
	return nil, nil
}

func (s *DatabaseStorage) Ping() error {
	return s.pool.Ping(context.Background())
}

type Server struct {
	Storage        models.Storable
	Cfg            models.Config
	App            *fiber.App
	ShortURLPrefix string
	Result         string `json:"URL"`
	Logger         *logrus.Logger
	DBPool         *pgxpool.Pool // пул соединений с базой данных
}

func NewServer(cfg models.Config, storage models.Storable) *Server {

	if cfg.FileStoragePath == "" {
		fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
		if fileStoragePath != "" {
			cfg.FileStoragePath = fileStoragePath
		} else {
			cfg.FileStoragePath = "/tmp/short-url-db.json"
		}
	}

	log := fiber.New()
	log.Use(logger.New(logger.Config{
		Output: &fiberLogger{logger: logrus.New()},
		Format: "{\"status\": ${status}, \"duration\": \"${latency}\", \"method\": \"${method}\", \"path\": \"${path}\", \"resp\": \"${resBody}\"}\n",
	}))

	logger := logrus.New()

	server := &Server{
		Cfg:            cfg,
		Storage:        storage,
		App:            log,
		ShortURLPrefix: cfg.BaseURL + "/",
		Logger:         logger,
	}

	server.setupRoutes()

	if _, err := os.Stat(cfg.FileStoragePath); !os.IsNotExist(err) {
		err := server.loadStorageFromFile(cfg.FileStoragePath)
		if err != nil {
			logger.Errorf("Failed to load storage from file: %v", err)
		}
	}

	return server
}

func (s *Server) setupRoutes() {
	s.App.Post("/api/shorten", s.shortenAPIHandler)
	s.App.Post("/", s.shortenURLHandler)
	s.App.Get("/ping", s.PingHandler)
	s.App.Get("/:id", s.redirectToOriginalURL)
}

func (s *Server) Run() error {
	s.setupRoutes()

	if s.Cfg.FileStoragePath != "" {
		err := s.saveStorageToFile(s.Cfg.FileStoragePath)
		if err != nil {
			s.Logger.Errorf("Failed to save storage to file: %v", err)
		}
	}

	return s.App.Listen(s.Cfg.Address)
}
