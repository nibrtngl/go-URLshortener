package server

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortenURLHandler(t *testing.T) {
	config := Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	server := NewServer(config)
	server.App.Post("/", server.shortenURLHandler)

	tests := []struct {
		name         string
		path         string
		expectedCode int
		URL          string
		Header       string
	}{
		// Existing tests...

		// New test for the new endpoint
		{
			name:         "get HTTP status 400 for invalid request body",
			path:         "/api/shorten",
			expectedCode: fiber.StatusBadRequest,
			URL:          `{"invalid": "request"}`, // Update the request body here
			Header:       "application/json",
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
		assert.Equalf(t, test.Header, resp.Header.Get("Content-type"), test.name)
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.name)

		resp.Body.Close()
	}
}

func TestRedirectToOriginalURL(t *testing.T) {
	config := Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	server := NewServer(config)
	server.App.Get("/:id", server.redirectToOriginalURL)

	tests := []struct {
		name         string
		path         string
		id           string
		expectedCode int
		URL          string
		Header       string
	}{
		// Existing tests...

		// New test for the new endpoint
		{
			name:         "HTTP status 307 for new endpoint",
			path:         "/short-id",
			id:           "short-id",
			expectedCode: http.StatusTemporaryRedirect,
			URL:          "http://example.com",
		},
	}

	server.Storage["short-id"] = "http://example.com"

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
