package server

import (
	"github.com/gofiber/fiber/v2"
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
	return &Server{
		Config:         config,
		Storage:        make(map[string]string),
		App:            fiber.New(),
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
