package models

import (
	"github.com/sirupsen/logrus"
)

// интерфейс, которому должно соответствовать хранилище
type Storable interface {
	userStorable
	GetAllKeys() ([]string, error)
}

// для таблиц user в бд
type userStorable interface {
	GetUrl(id string) (string, error)
	SetUrl(id, url string)
	Ping() error
}

type Storage struct {
	Data  data
	Users users
}

type data map[string]string
type users map[string][]string

type ErrorResponse struct {
	Error string `json:"error"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string
	DSN             string `env:"DATABASE_DSN" json:"database_dsn"`
}

type fiberLogger struct {
	logger *logrus.Logger
}
