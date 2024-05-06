package server

import (
	"context"
	"encoding/json"
	"errors"
	"fiber-apis/internal/models"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
)

type InternalStorage struct {
	urls map[string]string
}

func NewInternalStorage() *InternalStorage {
	return &InternalStorage{
		urls: make(map[string]string),
	}
}

func (s *InternalStorage) GetURL(shortURL string) (string, error) {
	originalURL, ok := s.urls[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return originalURL, nil
}

func (s *InternalStorage) SetURL(id, url string) error {
	s.urls[id] = url
	return nil
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

func (s *DatabaseStorage) GetURL(shortURL string) (string, error) {
	query := "SELECT original_url FROM urls WHERE short_url = $1"
	row := s.pool.QueryRow(context.Background(), query, shortURL)

	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("no original URL found for shortURL %s", shortURL)
		}
		return "", fmt.Errorf("failed to get URL from database: %v", err)
	}

	return originalURL, nil
}

func (s *DatabaseStorage) SetURL(id, url string) error {
	query := "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)"
	result, err := s.pool.Exec(context.Background(), query, id, url)
	if err != nil {
		return fmt.Errorf("failed to insert URL into database: %v", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows were inserted")
	}
	return nil
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

func (s *Server) shortenBatchURLHandler(c *fiber.Ctx) error {
	var req []models.BatchShortenRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		errResponse := models.ErrorResponse{
			Error: "bad request: Invalid json format",
		}
		return c.Status(http.StatusBadRequest).JSON(errResponse)
	}

	var resp []models.BatchShortenResponse
	for _, item := range req {
		if !isValidURL(item.OriginalURL) {
			return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
		}

		id := generateShortID()
		s.Storage.SetURL(id, item.OriginalURL)

		shortURL, _ := url.JoinPath(s.ShortURLPrefix, id)
		resp = append(resp, models.BatchShortenResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

func NewServer(cfg models.Config, pool *pgxpool.Pool) *Server {
	var storage models.Storable

	if cfg.DatabaseDSN != "" {
		storage = NewDatabaseStorage(pool)
	} else {
		storage = NewInternalStorage()
	}
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
	s.App.Post("/api/shorten/batch", s.shortenBatchURLHandler)
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
