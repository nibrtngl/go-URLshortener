package server

import (
	"encoding/json"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

const UserID = "userID"

func (s *Server) shortenURLHandler(c *fiber.Ctx) error {
	originalURL := c.Body()
	userID := c.Cookies(UserID)
	if s.CookieHandler == nil {
		s.CookieHandler = securecookie.New([]byte("very-secret"), []byte("a-lot-secret"))
	}
	if userID == "" || !s.Valid(userID) {
		userID = generateUserID()
		value := map[string]string{
			UserID: userID,
		}
		encoded, err := s.CookieHandler.Encode(UserID, value)
		if err == nil {
			c.Cookie(&fiber.Cookie{
				Name:     UserID,
				Value:    encoded,
				HTTPOnly: true,
			})
		}
	}

	if !isValidURL(string(originalURL)) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	}

	id := generateShortID()

	dbid, err := s.Storage.SetURL(id, string(originalURL), userID)
	c.Cookie(&fiber.Cookie{Name: UserID, Value: userID})
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
	userID := c.Cookies(UserID)
	urlData, err := s.Storage.GetURL(id, userID)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("404, not found")
	}
	if urlData.IsDeleted {
		return c.Status(http.StatusGone).SendString("410, gone")
	}
	if s.CookieHandler == nil {
		s.CookieHandler = securecookie.New([]byte("very-secret"), []byte("a-lot-secret"))
	}
	if userID == "" {
		value := map[string]string{
			UserID: "userID",
		}
		encoded, err := s.CookieHandler.Encode(UserID, value)
		if err == nil {
			c.Cookie(&fiber.Cookie{
				Name:     UserID,
				Value:    encoded,
				HTTPOnly: true,
			})
		}
	} else {
		c.Cookie(&fiber.Cookie{Name: UserID, Value: userID})
	}
	originalURL := urlData.OriginalURL // Access the OriginalURL field of the models.URL struct
	if !isValidURL(originalURL) {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Invalid URL format")
	} else {
		c.Set("Location", originalURL)
		return c.Status(http.StatusTemporaryRedirect).SendStatus(http.StatusTemporaryRedirect)
	}
}

func (s *Server) shortenAPIHandler(c *fiber.Ctx) error {
	var req models.ShortenRequest
	if s.CookieHandler == nil {
		s.CookieHandler = securecookie.New([]byte("very-secret"), []byte("a-lot-secret"))
	}
	userID := c.Cookies(UserID)
	if userID == "" {
		value := map[string]string{
			UserID: "userID",
		}
		encoded, err := s.CookieHandler.Encode(UserID, value)
		if err == nil {
			c.Cookie(&fiber.Cookie{
				Name:     UserID,
				Value:    encoded,
				HTTPOnly: true,
			})
		}
	} else {
		c.Cookie(&fiber.Cookie{Name: UserID, Value: userID})
	}
	c.Cookie(&fiber.Cookie{Name: UserID, Value: userID})
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
	dbid, err := s.Storage.SetURL(id, req.URL, c.Cookies(UserID))

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

func (s *Server) getUserURLsHandler(c *fiber.Ctx) error {
	userID := c.Cookies(UserID)
	if s.CookieHandler == nil {
		s.CookieHandler = securecookie.New([]byte("very-secret"), []byte("a-lot-secret"))
	}

	if userID == "" || !s.Valid(userID) {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized: Invalid user ID")
	}

	urls, err := s.Storage.GetUserURLs(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user URLs",
		})
	}

	if len(urls) == 0 {
		return c.Status(http.StatusNoContent).SendString("No Content: No URLs found for this user")
	}

	response := make([]models.RespPair, len(urls))
	for i, url := range urls {
		response[i] = models.RespPair{
			ShortURL:    s.ShortURLPrefix + url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
	}

	return c.Status(http.StatusOK).JSON(response)
}

func (s *Server) deleteURLsHandler(c *fiber.Ctx) error {
	var ids []string
	if err := json.Unmarshal(c.Body(), &ids); err != nil {
		errResponse := models.ErrorResponse{
			Error: "bad request: Invalid json format",
		}
		return c.Status(http.StatusBadRequest).JSON(errResponse)
	}

	userID := c.Cookies(UserID)
	if err := s.Storage.SetURLsAsDeleted(ids, userID); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(http.StatusAccepted).SendString("Accepted")
}

func (s *Server) shortenBatchURLHandler(c *fiber.Ctx) error {
	if s.CookieHandler == nil {
		s.CookieHandler = securecookie.New([]byte("very-secret"), []byte("a-lot-secret"))
	}
	var req []models.BatchShortenRequest
	userID := c.Cookies(UserID)
	if userID == "" {
		value := map[string]string{
			UserID: "userID",
		}
		encoded, err := s.CookieHandler.Encode(UserID, value)
		if err == nil {
			c.Cookie(&fiber.Cookie{
				Name:     UserID,
				Value:    encoded,
				HTTPOnly: true,
			})
		}
	} else {
		c.Cookie(&fiber.Cookie{Name: UserID, Value: userID})
	}
	c.Cookie(&fiber.Cookie{Name: UserID, Value: userID})
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
		s.Storage.SetURL(id, item.OriginalURL, c.Cookies(UserID))

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
