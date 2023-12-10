package server

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
)

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	originalURL := string(c.Body())
	if !isValidURL(originalURL) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()
	s.Storage[id] = originalURL

	shortURL, _ := url.JoinPath(s.ShortURLPrefix, id)

	return c.Status(http.StatusCreated).SendString(shortURL)
}

func (s *Server) redirectToOriginalURL(c *fiber.Ctx) error {
	id := c.Params("id")
	originalURL, exist := s.Storage[id]
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
