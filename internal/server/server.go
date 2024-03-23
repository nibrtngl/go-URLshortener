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
	"sync"
)

type MyStorage struct {
	data map[string]string
	db   *pgxpool.Pool
	mu   sync.RWMutex
}

func (s *MyStorage) InitDB(ctx context.Context, connString string) error {
	// Подключаемся к базе данных PostgreSQL с помощью pgx.
	db, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return err
	}

	// Проверяем, что соединение с базой данных установлено корректно.
	err = db.Ping(ctx)
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *MyStorage) GetURL(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	originalURL, ok := s.data[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

// SetURL сохраняет оригинальный URL и возвращает сокращенный.
func (s *MyStorage) SetURL(id, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = url
}

func (s *MyStorage) Ping() error {
	return nil
}

func (s *MyStorage) GetAllKeys() ([]string, error) {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
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

	// Исправлены ссылки на Config
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
	s.App.Get("/:id", s.redirectToOriginalURL)
	s.App.Get("/ping", s.pingHandler)
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
