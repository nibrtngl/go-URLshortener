package main

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_shortenURLHandler(t *testing.T) {
	type set struct {
		name         string
		path         string
		expectedCode int
		URL          string
	}

	app := fiber.New()

	urlMapping := make(map[string]string)

	app.Post("/", func(c *fiber.Ctx) error {
		originalURL := string(c.Body())
		id := generateShortID()
		urlMapping[id] = originalURL

		shortURL := "http://localhost:8080/" + id
		return c.Status(fiber.StatusCreated).SendString(shortURL)
	})

	tests := []set{
		{
			name:         "get HTTP status 201",
			path:         "/",
			expectedCode: http.StatusCreated,
			URL:          "https://example.com",
		},
	}

	for _, test := range tests {
		a := bytes.NewBuffer([]byte(test.URL))
		req := httptest.NewRequest(http.MethodPost, test.path, a)
		resp, err := app.Test(req, -1)
		if err != nil {
			log.Println(err)
			continue
		}
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.name)
		err = resp.Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func Test_redirectToOriginalURL(t *testing.T) {
	type set struct {
		name         string
		path         string
		expectedCode int
		URL          string
	}

	app := fiber.New()

	urlMapping := make(map[string]string)

	app.Post("/", func(c *fiber.Ctx) error {
		originalURL := string(c.Body())
		id := generateShortID()
		urlMapping[id] = originalURL

		shortURL := "http://localhost:8080/" + id
		return c.Status(fiber.StatusTemporaryRedirect).SendString(shortURL)
	})

	tests := []set{
		{
			name:         "HTTP status 307",
			path:         "/",
			expectedCode: http.StatusTemporaryRedirect,
			URL:          "https://example.com",
		},
	}

	for _, test := range tests {
		a := bytes.NewBuffer([]byte(test.URL))
		req := httptest.NewRequest(http.MethodPost, test.path, a)
		resp, err := app.Test(req, -1)
		if err != nil {
			log.Println(err)
			continue
		}
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.name)
		err = resp.Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}
}
