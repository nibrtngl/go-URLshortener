package server

import (
	"bytes"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortenURLHandler(t *testing.T) {
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
		name         string
		path         string
		expectedCode int
		URL          string
	}{
		{
			name:         "get HTTP status 201",
			path:         "/",
			expectedCode: http.StatusCreated,
			URL:          "https://example.com",
		},
		{
			name:         "get invalid URL",
			path:         "/",
			expectedCode: http.StatusBadRequest,
			URL:          "!_@O",
		},
		{
			name:         "get invalid path",
			path:         "/invalid_path123s",
			expectedCode: http.StatusMethodNotAllowed,
			URL:          "https://example.com",
		},
	}

	for _, test := range tests {
		b := bytes.NewBuffer([]byte(test.URL))
		req := httptest.NewRequest(http.MethodPost, test.path, b)

		resp, err := server.App.Test(req, -1)
		if err != nil {
			log.Println(err)
			continue
		}
		assert.Equalf(t, "text/plain; charset=utf-8", resp.Header.Get("Content-type"), test.name)
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.name)

		resp.Body.Close()
	}

}

func TestRedirectToOriginalURL(t *testing.T) {
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
		name         string
		path         string
		id           string
		expectedCode int
		URL          string
	}{
		{
			name:         "HTTP status 307",
			path:         "/1",
			id:           "1",
			expectedCode: http.StatusTemporaryRedirect,
			URL:          "http://yandex.ru",
		},
		{
			name:         "get invalid URL",
			path:         "/invalid_id",
			id:           "invalid_id",
			expectedCode: http.StatusBadRequest,
			URL:          "",
		},
		{
			name:         "get status not found",
			path:         "/invalid_id2",
			id:           "invalid_id2",
			expectedCode: http.StatusNotFound,
			URL:          "",
		},
	}
	server.Storage.(*MyStorage).data["invalid_id"] = "!$#09"
	server.Storage.(*MyStorage).data["1"] = "http://yandex.ru"

	for _, test := range tests {

		req := httptest.NewRequest(http.MethodGet, test.path, nil)

		resp, err := server.App.Test(req, -1)
		if err != nil {
			t.Errorf("Error testing %s: %s", test.name, err.Error())
			continue
		}
		assert.Equalf(t, test.URL, resp.Header.Get("Location"), "unexpected redirect URL")
		assert.Equalf(t, "text/plain; charset=utf-8", resp.Header.Get("Content-type"), test.name)
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.name)
		resp.Body.Close()

	}
}

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

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Error reading response body: %s", err)
			continue
		}

		assert.Equalf(t, test.expectedResult, string(body), test.name)

		resp.Body.Close()
	}
}
