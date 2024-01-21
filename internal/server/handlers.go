package server

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
)

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	request := new(ShortenRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Проверка действительности URL-адреса
	_, err := url.ParseRequestURI(request.URL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	// Создание сокращенного URL и сохранение его в хранилище
	shortURL := s.ShortURLPrefix + generateShortID()
	s.Storage[shortURL] = request.URL

	response := ShortenResponse{
		Result: shortURL,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
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
