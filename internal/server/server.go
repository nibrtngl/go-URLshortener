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
