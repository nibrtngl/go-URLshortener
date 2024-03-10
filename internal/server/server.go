package server

import (
	"fiber-apis/internal/models"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sirupsen/logrus"
	"os"
)

type MyStorage struct {
	data map[string]string
}

func (s *MyStorage) GetURL(id string) (string, error) {
	url, ok := s.data[id]
	if !ok {
		return "", fmt.Errorf("URL not found for ID: %s", id)
	}
	return url, nil
}

func (s *MyStorage) SetURL(id, url string) {
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
}

func NewServer(cfg models.Config) *Server {

	storage := &MyStorage{
		data: make(map[string]string),
	}
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
		Cfg:            cfg,     // Исправлена ссылка на Config
		Storage:        storage, // Заменено на конкретную реализацию хранилища
		App:            log,
		ShortURLPrefix: cfg.BaseURL + "/", // Исправлено на Cfg.BaseURL
		Logger:         logger,
	}

	server.setupRoutes()

	// При запуске сервера проверяем, есть ли файл для загрузки данных
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
	s.App.Get("/ping", s.pingHandler) // Добавлен обработчик для маршрута /ping
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
