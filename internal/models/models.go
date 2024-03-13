package models

import (
	"github.com/sirupsen/logrus"
)

// интерфейс, которому должно соответствовать хранилище
type Storable interface {
	GetURL(id string) (string, error)
	SetURL(id, url string)
	GetAllKeys() ([]string, error)
	Ping() error
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

var Logger *logrus.Logger

type Config struct {
	Address         string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./tmp/short-url-db.json"`
	DATABASE_DSN    string `env:"DATABASE_DSN" envDefault:"host=localhost port=5432 dbname=url-db user=postgres connect_timeout=10 sslmode=prefer"`
}
