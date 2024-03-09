package server

import (
	"bytes"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortenAPIHandler(t *testing.T) {
	config := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
	server := &Server{
		Storage:        &MyStorage{data: make(map[string]string)},
		Cfg:            config,
		App:            fiber.New(),
		ShortURLPrefix: config.BaseURL + "/",
		Logger:         logrus.New(),
	}
	server.setupRoutes()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedResult string
	}{
		{
			name:           "Valid request",
			requestBody:    `{"url": "https://example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedResult: `{"result": "http://shorturl.com/abc123"}`,
		},
		{
			name:           "Invalid JSON format",
			requestBody:    `{"url123": "23https://example.com",}`,
			expectedStatus: http.StatusBadRequest,
			expectedResult: "Bad Request: Invalid JSON format",
		},
	}

	for _, test := range tests {
		b := bytes.NewBuffer([]byte(test.requestBody))
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", b)

		resp, err := server.App.Test(req, -1)
		if err != nil {
			log.Println(err)
			continue
		}
		assert.Equalf(t, "application/json", resp.Header.Get("Content-type"), test.name)
		assert.Equalf(t, test.expectedStatus, resp.StatusCode, test.name)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Error reading response body: %s", err)
			continue
		}

		assert.Equalf(t, test.expectedResult, string(body), test.name)

		resp.Body.Close()
	}
}
