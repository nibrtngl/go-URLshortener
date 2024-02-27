package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sirupsen/logrus"
	"os"
)

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
}

type Server struct {
	Config         Config
	Storage        map[string]string
	App            *fiber.App
	ShortURLPrefix string
	Result         string         `json:"URL"`
	Logger         *logrus.Logger // Add a logger field to the Server struct
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
		Output: &fiberLogger{logger: logrus.New()}, // Set the output to the custom fiberLogger
		Format: "{\"status\": ${status}, \"duration\": \"${latency}\", \"method\": \"${method}\", \"path\": \"${path}\", \"resp\": \"${resBody}\"}\n",
	}))

	logger := logrus.New()

	server := &Server{
		Config:         config,
		Storage:        make(map[string]string),
		App:            log,
		ShortURLPrefix: config.BaseURL + "/",
		Logger:         logger, // Assign the logger to the Server struct
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	s.App.Post("/api/shorten", s.shortenAPIHandler)
	s.App.Post("/", s.shortenURLHandler)
	s.App.Get("/:id", s.redirectToOriginalURL)
}

func (s *Server) Run() error {
	s.setupRoutes()

	if s.Config.FileStoragePath != "" {
		err := s.saveStorageToFile(s.Config.FileStoragePath)
		if err != nil {
			logrus.Errorf("Failed to save storage to file: %v", err)
		}
	}

	return s.App.Listen(s.Config.Address)
}
