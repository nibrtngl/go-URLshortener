package models

type Storable interface {
	GetURL(id string) (string, error)
	SetURL(id, url string)
	GetAllKeys() ([]string, error) // Добавлен новый метод для получения всех ключей
}

// интерфейс, которому должно соответствовать хранилище

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
