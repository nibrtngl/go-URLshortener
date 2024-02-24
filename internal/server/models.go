package server

import (
	"flag"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
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
	FileStoragePath *string
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
	log := fiber.New()
	log.Use(logger.New(logger.Config{
		Output: &fiberLogger{logger: logrus.New()}, // Set the output to the custom fiberLogger
		Format: "{\"status\": ${status}, \"duration\": \"${latency}\", \"method\": \"${method}\", \"path\": \"${path}\", \"resp\": \"${resBody}\"}\n",
	}))
	log.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	logger := logrus.New()

	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePath == "" {
		fileStoragePathFlag := flag.String("f", "/tmp/short-url-db.json", "Path to the file for storing data")
		fileStoragePath = *fileStoragePathFlag // Dereference the pointer to get the string value
	}
	config.FileStoragePath = &fileStoragePath // Assign the address of the string to FileStoragePath

	server := &Server{
		Config:         config,
		Storage:        make(map[string]string),
		App:            log,
		ShortURLPrefix: config.BaseURL + "/",
		Logger:         logger, // Assign the logger to the Server struct
	}

	if *config.FileStoragePath != "" {
		// Загрузка данных из файла
		err := server.loadStorageFromFile(*config.FileStoragePath)
		if err != nil {
			logrus.Errorf("Failed to load storage from file: %v", err)
		}
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
	return s.App.Listen(s.Config.Address)
}
