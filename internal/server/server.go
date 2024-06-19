package server

import (
	"encoding/json"
	"fiber-apis/internal/db"
	"fiber-apis/internal/localstorage"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gorilla/securecookie"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Storable interface {
	GetURL(shortURL string, userID string) (models.URL, error)
	SetURL(id, url string, userID string) (string, error)
	SetURLsAsDeleted(ids []string, userID string) error
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
	CookieHandler  *securecookie.SecureCookie
}

func NewServer(cfg models.Config, pool *pgxpool.Pool, cookieHandler *securecookie.SecureCookie) *Server {
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
		CookieHandler:  cookieHandler,
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

func (s *Server) deleteURLsHandler(c *fiber.Ctx) error {
	var ids []string
	if err := json.Unmarshal(c.Body(), &ids); err != nil {
		errResponse := models.ErrorResponse{
			Error: "bad request: Invalid json format",
		}
		return c.Status(http.StatusBadRequest).JSON(errResponse)
	}

	userID := c.Cookies(UserID)
	if err := s.Storage.SetURLsAsDeleted(ids, userID); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(http.StatusAccepted).SendString("Accepted")
}

func (s *Server) Valid(userID string) bool {
	return userID != ""
}

func (s *Server) setupRoutes() {
	s.App.Post("/api/shorten", s.shortenAPIHandler)
	s.App.Post("/", s.shortenURLHandler)
	s.App.Get("/ping", s.PingHandler)
	s.App.Get("/:id", s.redirectToOriginalURL)
	s.App.Post("/api/shorten/batch", s.shortenBatchURLHandler)
	s.App.Get("/api/user/urls", s.getUserURLsHandler)
	s.App.Delete("/api/user/urls", s.deleteURLsHandler)
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
