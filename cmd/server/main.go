package main

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"net/http"
	"net/url"
)

type Config struct {
	Address string
	BaseURL string
}

type Server struct {
	config         Config
	storage        map[string]string
	app            *fiber.App
	ShortURLPrefix string
}

func NewServer(config Config) *Server {
	return &Server{
		config:         config,
		storage:        make(map[string]string),
		app:            fiber.New(),
		ShortURLPrefix: config.BaseURL + "/",
	}
}

func (s *Server) setupRoutes() {
	s.app.Post("/", s.shortenURLHandler)
	s.app.Get("/:id", s.redirectToOriginalURL)
}

func (s *Server) Run() error {
	s.setupRoutes()
	return s.app.Listen(s.config.Address)
}

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	originalURL := string(c.Body())
	if !isValidURL(originalURL) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()
	s.storage[id] = originalURL

	shortURL := fmt.Sprintf("%s%s", s.ShortURLPrefix, id)

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
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY0123456789"
	idLength := 8
	b := make([]byte, idLength)

	// Generate a unique identifier
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func main() {
	// Parse command-line arguments
	address := flag.String("a", "localhost:8080", "address to run the HTTP server")
	baseURL := flag.String("b", "http://localhost:8080", "base URL for the shortened URL")
	flag.Parse()

	// Initialize configuration
	config := Config{
		Address: *address,
		BaseURL: *baseURL,
	}

	// Initialize server with the configuration
	server := NewServer(config)

	// Run the server
	err := server.Run()
	if err != nil {
		fmt.Printf("Error running server: %v", err)
	}
}
