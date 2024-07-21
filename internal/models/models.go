package models

import (
	"github.com/sirupsen/logrus"
)

// URL представляет собой структуру, которая представляет URL.
type URL struct {
	ShortURL    string
	OriginalURL string
	IsDeleted   bool
}

// RespPair представляет собой структуру, которая представляет пару ответов.
type RespPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchShortenRequest представляет собой структуру, которая представляет запрос на сокращение URL в пакетном режиме.
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchShortenResponse представляет собой структуру, которая представляет ответ на запрос на сокращение URL
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ErrorResponse представляет собой структуру, которая представляет ответ об ошибке.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ShortenResponse представляет собой структуру, которая представляет ответ на запрос на сокращение URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortenRequest представляет собой структуру, которая представляет запрос на сокращение URL.
type ShortenRequest struct {
	URL string `json:"url"`
}

// Logger представляет собой переменную для логирования.
var Logger *logrus.Logger

// Config представляет собой структуру, которая представляет конфигурацию.
type Config struct {
	Address         string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./tmp/short-url-db.json"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`
}
