package server

import (
	"encoding/json"
	"fiber-apis/internal/models"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	originalURL := c.Body()
	if !isValidURL(string(originalURL)) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()
	s.Storage.SetURL(id, string(originalURL))

	err := s.saveStorageToFile(s.Cfg.FileStoragePath)
	if err != nil {
		logrus.Errorf("Failed to save storage to file: %v", err)
	}

	shortURL, _ := url.JoinPath(s.ShortURLPrefix, id)

	return c.Status(http.StatusCreated).SendString(shortURL)
}

func (s *Server) redirectToOriginalURL(c *fiber.Ctx) error {
	id := c.Params("id")
	originalURL, err := s.Storage.GetURL(id)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("404, not found")
	}
	if !isValidURL(originalURL) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	} else {
		c.Set("Location", originalURL)
		return c.Status(http.StatusTemporaryRedirect).SendStatus(http.StatusTemporaryRedirect)
	}
}

func (s *Server) shortenAPIHandler(c *fiber.Ctx) error {
	var req models.ShortenRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		errResponse := models.ErrorResponse{
			Error: "bad request: Invalid json format",
		}
		return c.Status(http.StatusBadRequest).JSON(errResponse)
	}

	if !isValidURL(req.URL) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()
	s.Storage.SetURL(id, req.URL)

	shortURL, _ := url.JoinPath(s.ShortURLPrefix, id)

	resp := models.ShortenResponse{
		Result: shortURL,
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

func (s *Server) pingHandler(c *fiber.Ctx) error {
	err := s.Storage.Ping()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to ping the database: %v", err))
	}
	return c.SendStatus(fiber.StatusOK)
}
