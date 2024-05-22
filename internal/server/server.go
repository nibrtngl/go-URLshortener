package server

import (
	"encoding/json"
	"fiber-apis/internal/db"
	"fiber-apis/internal/localstorage"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
)

type Storable interface {
	GetURL(id string) (string, error)
	SetURL(id, url string) error
	GetAllKeys() ([]string, error)
	Ping() error
}

type Server struct {
	Storage        Storable
	Cfg            models.Config
	App            *fiber.App
	ShortURLPrefix string
	Result         string `json:"URL"`
	Logger         *logrus.Logger
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
	var storage Storable

	if cfg.DatabaseDSN != "" {
		storage = db.NewDatabaseStorage(pool)
	} else {
		storage = localstorage.NewInternalStorage()
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
		err := s.SaveStorageToFile(s.Cfg.FileStoragePath)
		if err != nil {
			s.Logger.Errorf("Failed to save storage to file: %v", err)
		}
	}

	return s.App.Listen(s.Cfg.Address)
}
