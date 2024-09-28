package server

import (
	"fiber-apis/internal/db"
	"fiber-apis/internal/localstorage"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gorilla/securecookie"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net"
	"os"
)

// Storable представляет интерфейс для хранилища URL.
type Storable interface {
	GetURL(shortURL string, userID string) (models.URL, error)
	SetURL(id, url string, userID string) (string, error)
	SetURLsAsDeleted(ids []string, userID string) error
	GetAllKeys() ([]string, error)
	GetUserURLs(userID string) ([]models.URL, error)
	Ping() error
	SaveToFile(filePath string) error
	LoadFromFile(filePath string) error
	GetURLsCount() (int, error)
	GetUsersCount() (int, error)
}

// Server представляет структуру сервера.
type Server struct {
	Storage        Storable
	Cfg            models.Config
	App            *fiber.App
	ShortURLPrefix string
	Result         string `json:"URL"`
	Logger         *logrus.Logger
	CookieHandler  *securecookie.SecureCookie
}

// NewServer создает новый экземпляр сервера.
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

	// Загрузка данных из файла
	if _, err := os.Stat(cfg.FileStoragePath); !os.IsNotExist(err) {
		err := server.Storage.LoadFromFile(cfg.FileStoragePath)
		if err != nil {
			logger.Errorf("Failed to load storage from file: %v", err)
		}
	}

	server.setupRoutes()

	return server
}

func (s *Server) StatsHandler(c *fiber.Ctx) error {
	// Проверка на доверенный IP
	realIP := c.Get("X-Real-IP")
	if !s.isIPTrusted(realIP) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Получение статистики
	urlsCount, err := s.Storage.GetURLsCount()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stats"})
	}

	usersCount, err := s.Storage.GetUsersCount()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stats"})
	}

	return c.JSON(fiber.Map{
		"urls":  urlsCount,
		"users": usersCount,
	})
}

// Valid проверяет, является ли пользователь действительным.
func (s *Server) Valid(userID string) bool {
	return userID != ""
}

func (s *Server) SaveData(path string) error {
	if path != "" {
		err := s.Storage.SaveToFile(path)
		if err != nil {
			s.Logger.Errorf("Failed to save storage to file: %v", err)
			return err
		}
		s.Logger.Infof("Data successfully saved to %s", path)
	}
	return nil
}

func (s *Server) isIPTrusted(ip string) bool {
	if s.Cfg.TrustedSubnet == "" {
		return false
	}
	_, subnet, err := net.ParseCIDR(s.Cfg.TrustedSubnet)
	if err != nil {
		return false
	}
	clientIP := net.ParseIP(ip)
	return subnet.Contains(clientIP)
}

// setupServerForTesting тестирует сервер.
func setupServerForTesting() *Server {
	cfg := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	storage := localstorage.NewInternalStorage()

	server := &Server{
		Storage:        storage,
		Cfg:            cfg,
		App:            fiber.New(),
		ShortURLPrefix: cfg.BaseURL + "/",
		Logger:         logrus.New(),
	}

	server.setupRoutes()

	return server
}

// setupRoutes настраивает маршруты.
func (s *Server) setupRoutes() {
	s.App.Post("/api/shorten", s.ShortenAPIHandler)
	s.App.Post("/", s.ShortenURLHandler)
	s.App.Get("/ping", s.PingHandler)
	s.App.Get("/:id", s.RedirectToOriginalURL)
	s.App.Post("/api/shorten/batch", s.ShortenBatchURLHandler)
	s.App.Get("/api/user/urls", s.GetUserURLsHandler)
	s.App.Delete("/api/user/urls", s.DeleteURLsHandler)
	s.App.Get("/api/internal/stats", s.StatsHandler)

}

func (s *Server) RunTLS(certFile, keyFile string) error {
	return s.App.ListenTLS(s.Cfg.Address, certFile, keyFile)
}

// Run запускает сервер.
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
