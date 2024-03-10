package models

import "github.com/sirupsen/logrus"

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

var (
	DBDSN      string
	ListenAddr string
)

var Logger *logrus.Logger

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string
	DSN             string `env:"DATABASE_DSN" json:"database_dsn"`
}
