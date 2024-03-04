package models

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"os"
)

type Storable interface {
	urlsStorable
}

type urlsStorable interface {
	Get(key string) (string, error)
	Set(val, pth string) (string, error)
	Ping() error
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string
	DSN             string `env:"DATABASE_DSN"`
}

type Server struct {
	Config         Config
	Storage        models.Config
	App            *fiber.App
	ShortURLPrefix string
	Result         string `json:"URL"`
	Logger         *logrus.Logger
	DB             *pgx.Conn
}

type fiberLogger struct {
	logger *logrus.Logger
}

func (fl *fiberLogger) Write(p []byte) (n int, err error) {
	fl.logger.Info(string(p))
	return len(p), nil
}

func NewServer(config Config) *Server {
	if config.FileStoragePath == "" {
		fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
		if fileStoragePath != "" {
			config.FileStoragePath = fileStoragePath
		} else {
			config.FileStoragePath = "/tmp/short-url-db.json"
		}
	}

	log := fiber.New()
	log.Use(logger.New(logger.Config{
		Output: &fiberLogger{logger: logrus.New()},
		Format: "{\"status\": ${status}, \"duration\": \"${latency}\", \"method\": \"${method}\", \"path\": \"${path}\", \"resp\": \"${resBody}\"}\n",
	}))

	logger := logrus.New()

	server := &Server{
		Config:         config,
		Storage:        make(map[string]string),
		App:            log,
		ShortURLPrefix: config.BaseURL + "/",
		Logger:         logger,
		DSN:            os.Getenv("DATABASE_DSN"), // Получаем строку подключения к БД из переменной окружения
	}

	// Подключаемся к БД
	db, err := pgx.Connect(context.Background(), server.DSN)
	if err != nil {
		server.Logger.Fatalf("Failed to connect to database: %v", err)
	}
	server.DB = db

	server.setupRoutes()

	// При запуске сервера проверяем, есть ли файл для загрузки данных
	if _, err := os.Stat(config.FileStoragePath); !os.IsNotExist(err) {
		err := server.loadStorageFromFile(config.FileStoragePath)
		if err != nil {
			server.Logger.Errorf("Failed to load storage from file: %v", err)
		}
	}

	return server
}

func (s *Server) setupRoutes() {
	s.App.Post("/api/shorten", s.shortenAPIHandler)
	s.App.Post("/", s.shortenURLHandler)
	s.App.Get("/:id", s.redirectToOriginalURL)
	s.App.Get("/ping", s.pingHandler) // Добавляем маршрут для проверки соединения с БД
}

func (s *Server) Run() error {
	s.setupRoutes()

	if s.Config.FileStoragePath != "" {
		err := s.saveStorageToFile(s.Config.FileStoragePath)
		if err != nil {
			s.Logger.Errorf("Failed to save storage to file: %v", err)
		}
	}

	return s.App.Listen(s.Config.Address)
}
