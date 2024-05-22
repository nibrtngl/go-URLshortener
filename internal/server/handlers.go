package server

import (
	"encoding/json"
	"errors"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
	"net/http"
	"net/url"
)

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	originalURL := c.Body()
	if !isValidURL(string(originalURL)) {
		s.Logger.Error("Bad Request: Invalid URL format")
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}
	id, err := s.Storage.SetURL(generateShortID(), string(originalURL))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.Logger.Warn("Conflict: URL already exists")
			return c.Status(http.StatusConflict).SendString("Conflict: URL already exists")
		}
		s.Logger.Error("Internal Server Error: ", err)
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
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
		s.Logger.Error("Bad Request: Invalid json format")
		errResponse := models.ErrorResponse{
			Error: "bad request: Invalid json format",
		}
		return c.Status(http.StatusBadRequest).JSON(errResponse)
	}

	if !isValidURL(req.URL) {
		s.Logger.Error("Bad Request: Invalid URL format")
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	return nil
}

func (s *Server) PingHandler(c *fiber.Ctx) error {
	err := s.Storage.Ping()
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to ping database")
	}
	return c.Status(http.StatusOK).SendString("Database connected")
}
