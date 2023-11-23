package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"net/http"
)

var urlMapping map[string]string

func main() {
	urlMapping = make(map[string]string)

	app := fiber.New()

	app.Post("/", shortenURLHandler)
	app.Get("/:id", redirectToOriginalURL)

	app.Listen(":8080")
}

func shortenURLHandler(c *fiber.Ctx) error {
	originalURL := string(c.Body())
	id := generateShortID()
	urlMapping[id] = originalURL

	shortURL := fmt.Sprintf("http://localhost:8080/%s", id)

	// Установка заголовка Location с оригинальным URL
	c.Set("Location", originalURL)

	return c.Status(http.StatusCreated).SendString(shortURL)
}

func redirectToOriginalURL(c *fiber.Ctx) error {
	id := c.Params("id")
	originalURL, exists := urlMapping[id]
	if !exists {
		return c.Status(http.StatusBadRequest).SendString("Bad Request")
	}

	// Установка заголовка Location с оригинальным URL
	c.Set("Location", originalURL)

	return c.Status(http.StatusTemporaryRedirect).SendString("")
}

func generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength := 8
	b := make([]byte, idLength)

	// Генерация уникального идентификатора
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
