package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Config struct {
	Address string
	BaseURL string
}

type Server struct {
	Config         Config
	Storage        map[string]string
	App            *fiber.App
	ShortURLPrefix string
}

func NewServer(config Config) *Server {
	log := fiber.New()
	log.Use(logger.New(logger.Config{
		Format: "{\"status\": ${status}, \"duration\": \"${latency}\", \"method\": \"${method}\", \"path\": \"${path}\", \"resp\": \"${resBody}\"}\n",
	}))
	return &Server{
		Config:         config,
		Storage:        make(map[string]string),
		App:            log,
		ShortURLPrefix: config.BaseURL + "/",
	}
}

func (s *Server) setupRoutes() {
	s.App.Post("/", s.shortenURLHandler)
	s.App.Get("/:id", s.redirectToOriginalURL)
}

func (s *Server) Run() error {
	s.setupRoutes()
	return s.App.Listen(s.Config.Address)
}
