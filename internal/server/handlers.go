package server

import (
	"encoding/json"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	originalURL := c.Body()
	userID := c.Cookies("userID")

	if !isValidURL(string(originalURL)) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()

	dbid, err := s.Storage.SetURL(id, string(originalURL), userID)
	c.Cookie(&fiber.Cookie{Name: "userID", Value: userID})
	shortURL, _ := url.JoinPath(s.ShortURLPrefix, dbid)
	if err != nil {
		logrus.Errorf("Failed to save url: %v", err)
	}
	if dbid != id {
		return c.Status(http.StatusConflict).SendString(shortURL)
	}
	err = s.saveStorageToFile(s.Cfg.FileStoragePath)
	if err != nil {
		logrus.Errorf("Failed to save storage to file: %v", err)
	}

	return c.Status(http.StatusCreated).SendString(shortURL)
}

func (s *Server) redirectToOriginalURL(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Cookies("userID")
	originalURL, err := s.Storage.GetURL(id, userID)
	c.Cookie(&fiber.Cookie{Name: "userID", Value: userID})
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
	dbid, err := s.Storage.SetURL(id, req.URL, c.Cookies("userID"))

	shortURL, _ := url.JoinPath(s.ShortURLPrefix, dbid)

	resp := models.ShortenResponse{
		Result: shortURL,
	}

	if dbid != id {
		return c.Status(http.StatusConflict).JSON(resp)
	}
	if err != nil {
		errResponse := models.ErrorResponse{
			Error: err.Error(),
		}
		return c.Status(http.StatusBadRequest).JSON(errResponse)
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

func (s *Server) shortenBatchURLHandler(c *fiber.Ctx) error {
	var req []models.BatchShortenRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		errResponse := models.ErrorResponse{
			Error: "bad request: Invalid json format",
		}
		return c.Status(http.StatusBadRequest).JSON(errResponse)
	}

	var resp []models.BatchShortenResponse
	for _, item := range req {
		if !isValidURL(item.OriginalURL) {
			return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
		}

		id := generateShortID()
		s.Storage.SetURL(id, item.OriginalURL, c.Cookies("userID"))

		shortURL, _ := url.JoinPath(s.ShortURLPrefix, id)
		resp = append(resp, models.BatchShortenResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
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
