package server

import (
	"encoding/json"
	"fiber-apis/internal/db"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	originalURL := c.Body()
	if !isValidURL(string(originalURL)) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()
	shortURL, err := s.Storage.SetURL(id, string(originalURL))

	if err != nil {
		if err == db.ErrURLAlreadyExists {
			return c.Status(http.StatusConflict).SendString(shortURL)
		}
		logrus.Errorf("Failed to save URL to storage: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}

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
	shortURL, err := s.Storage.SetURL(id, req.URL)

	if err != nil {
		if err == db.ErrURLAlreadyExists {
			resp := models.ShortenResponse{
				Result: shortURL,
			}
			return c.Status(http.StatusConflict).JSON(resp)
		}
		logrus.Errorf("Failed to save URL to storage: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}

	resp := models.ShortenResponse{
		Result: shortURL,
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

func (s *Server) PingHandler(c *fiber.Ctx) error {
	err := s.Storage.Ping()
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to ping database")
	}
	return c.Status(http.StatusOK).SendString("Database connected")
}
