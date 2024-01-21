package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type Config struct {
	Address string
	BaseURL string
}

type Server struct {
	Config         Config
	Storage        map[string]string
	App            *fiber.App
	ShortURLPrefix string
	Result         string `json:"URL"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func NewServer(config Config) *Server {
	log := fiber.New()
	log.Use(logger.New(logger.Config{
		Format: "{\"status\": ${status}, \"duration\": \"${latency}\", \"method\": \"${method}\", \"path\": \"${path}\", \"resp\": \"${resBody}\"}\n",
	}))

	server := &Server{
		Config:         config,
		Storage:        make(map[string]string),
		App:            log,
		ShortURLPrefix: config.BaseURL + "/",
	}

	server.setupRoutes() // Вызов метода setupRoutes

	return server
}

func (s *Server) setupRoutes() {
	s.App.Post("/", s.shortenURLHandler)
	s.App.Get("/:id", s.redirectToOriginalURL)
	s.App.Post("/api/shorten", s.shortenURLHandler)
}

func (s *Server) Run() error {
	s.setupRoutes()
	return s.App.Listen(s.Config.Address)
}
