package server

import (
	"fiber-apis/internal/db"
	"fiber-apis/internal/localstorage"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Storable interface {
	GetURL(shortURL string, userID string) (string, error)
	SetURL(id, url string, userID string) (string, error)
	GetAllKeys() ([]string, error)
	GetUserURLs(userID string) ([]models.URL, error)
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
func (s *Server) getUserURLsHandler(c *fiber.Ctx) error {
	userID := c.Cookies("userID")

	// Если кука не содержит ID пользователя, возвращаем HTTP-статус 401 Unauthorized
	if userID == "" {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized: User ID not found in cookies")
	}

	urls, err := s.Storage.GetUserURLs(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user URLs",
		})
	}

	// Если список URL пуст, возвращаем HTTP-статус 204 No Content
	if len(urls) == 0 {
		return c.Status(http.StatusNoContent).SendString("No Content: No URLs found for this user")
	}

	response := make([]models.RespPair, len(urls))
	for i, url := range urls {
		response[i] = models.RespPair{
			ShortURL:    s.ShortURLPrefix + url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
	}

	return c.Status(http.StatusOK).JSON(response)
}

func (s *Server) setupRoutes() {
	s.App.Post("/api/shorten", s.shortenAPIHandler)
	s.App.Post("/", s.shortenURLHandler)
	s.App.Get("/ping", s.PingHandler)
	s.App.Get("/:id", s.redirectToOriginalURL)
	s.App.Post("/api/shorten/batch", s.shortenBatchURLHandler)
	s.App.Get("/api/user/urls", func(c *fiber.Ctx) error {
		return s.getUserURLsHandler(c)
	})
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
