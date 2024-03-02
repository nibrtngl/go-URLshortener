package server

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
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

	err := s.saveStorageToFile(s.Config.FileStoragePath)
	if err != nil {
		logrus.Errorf("Failed to save storage to file: %v", err)
	}

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

func (s *Server) shortenAPIHandler(c *fiber.Ctx) error {
	var req ShortenRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		errResponse := ErrorResponse{
			Error: "bad request: Invalid json format",
		}
		return c.Status(http.StatusBadRequest).JSON(errResponse)
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

	return c.Status(http.StatusCreated).JSON(resp)
}

func (s *Server) pingHandler(c *fiber.Ctx) error {
	// Проверяем соединение с БД
	err := s.DB.Ping(context.Background())
	if err != nil {
		s.Logger.Println("Database connection error:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	return c.SendStatus(fiber.StatusOK)
}
