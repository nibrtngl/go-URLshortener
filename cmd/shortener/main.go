package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/rand"
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
	return c.Status(fiber.StatusCreated).SendString(shortURL)
}

func redirectToOriginalURL(c *fiber.Ctx) error {
	id := c.Params("id")
	originalURL, exists := urlMapping[id]
	if !exists {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	response := fmt.Sprintf(originalURL)
	return c.Status(fiber.StatusTemporaryRedirect).SendString(response)
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
