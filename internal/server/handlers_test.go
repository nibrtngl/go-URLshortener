package server

import (
	"bytes"
	"encoding/json"
	"fiber-apis/internal/localstorage"
	"fiber-apis/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestShortenURLHandler(t *testing.T) {
	config := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	server := &Server{
		Storage:        localstorage.NewInternalStorage(),
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
		Storage:        localstorage.NewInternalStorage(),
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
	server.Storage.SetURL("invalid_id", "!$#09", "1")
	server.Storage.SetURL("1", "http://yandex.ru", "1")

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
		BaseURL: "http://localhost:8080", //
	}
	server := &Server{
		Storage:        localstorage.NewInternalStorage(),
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
	}{
		{
			name:           "Valid request",
			requestBody:    `{"url": "https://example.com"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON format",
			requestBody:    `{"url123": "23https://example.com",}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		b := bytes.NewBuffer([]byte(test.requestBody))
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", b)

		resp, err := server.App.Test(req, -1)
		if err != nil {
			t.Errorf("Error testing %s: %s", test.name, err.Error())
			continue
		}

		assert.Equalf(t, "application/json", resp.Header.Get("Content-type"), test.name)
		assert.Equalf(t, test.expectedStatus, resp.StatusCode, test.name)

		if test.expectedStatus == http.StatusCreated {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Error reading response body: %s", err)
				continue
			}

			var result map[string]string
			if err := json.Unmarshal(body, &result); err != nil {
				t.Errorf("Error unmarshalling JSON: %s", err)
				continue
			}

			shortURL, ok := result["result"]
			if !ok {
				t.Errorf("Expected 'result' field in response body, got: %v", result)
				continue
			}
			defer resp.Body.Close()

			expectedURL, _ := url.JoinPath(shortURL)
			assert.Equalf(t, expectedURL, shortURL, "Expected shortened URL does not match")
		}
	}
}

func TestGetUserURLsHandler(t *testing.T) {

	server := NewServer(models.Config{}, nil, nil)
	server.Storage = localstorage.NewInternalStorage()

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.Header.Set("Cookie", "userID=userID")

	resp, err := server.App.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteURLsHandler(t *testing.T) {

	server := NewServer(models.Config{}, nil, nil)
	server.Storage = localstorage.NewInternalStorage()

	b := bytes.NewBuffer([]byte(`["url1", "url2"]`))
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", b)
	req.Header.Set("Cookie", "userID=userID")

	resp, err := server.App.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestShortenBatchURLHandler(t *testing.T) {

	server := NewServer(models.Config{}, nil, nil)
	server.Storage = localstorage.NewInternalStorage()

	b := bytes.NewBuffer([]byte(`[{"correlation_id": "1", "original_url": "https://example.com"}]`))
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", b)
	req.Header.Set("Cookie", "userID=userID")

	resp, err := server.App.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestPingHandler(t *testing.T) {

	server := NewServer(models.Config{}, nil, nil)
	server.Storage = localstorage.NewInternalStorage()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	resp, err := server.App.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
