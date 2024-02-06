package server

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
)

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	c.Type("application/json")
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
	c.Type("application/json")
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

func (s *Server) shortenAPIHandler(c *fiber.Ctx) error {
	var req ShortenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid JSON format")
	}

	if !isValidURL(req.URL) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()
	s.Storage[id] = req.URL

	shortURL, _ := url.JoinPath(s.ShortURLPrefix, id)

	resp := ShortenResponse{
		Result: shortURL,
	}

	return c.Status(http.StatusCreated).Type("application/json").JSON(resp)
}
