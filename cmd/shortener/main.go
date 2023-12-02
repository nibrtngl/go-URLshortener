package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"net/http"
	"net/url"
)

type Server struct {
	storage map[string]string
	app     *fiber.App
}

func NewServer() *Server {
	return &Server{
		storage: make(map[string]string),
		app:     fiber.New(),
	}
}

func (s *Server) setupRoutes() {
	s.app.Post("/", s.shortenURLHandler)
	s.app.Get("/:id", s.redirectToOriginalURL)
}

func (s *Server) Run(addr string) error {
	s.setupRoutes()
	return s.app.Listen(addr)
}

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	originalURL := string(c.Body())
	if !isValidURL(originalURL) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()
	s.storage[id] = originalURL

	shortURL := fmt.Sprintf("http://localhost:8080/%s", id)

	return c.Status(http.StatusCreated).SendString(shortURL)
}

func (s *Server) redirectToOriginalURL(c *fiber.Ctx) error {
	id := c.Params("id")
	originalURL, exist := s.storage[id]
	if !exist {
		return c.Status(http.StatusNotFound).SendString("404, not found")
	}
	if !isValidURL(originalURL) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	} else {
		c.Set("Location", originalURL)
		return c.SendStatus(http.StatusTemporaryRedirect)
	}
}

func isValidURL(url1 string) bool {
	_, err := url.ParseRequestURI(url1)
	return err == nil
}

func generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength := 8
	b := make([]byte, idLength)

	// Генерация уникального идентификатора
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func main() {
	server := NewServer()
	server.Run(":8080")
}
