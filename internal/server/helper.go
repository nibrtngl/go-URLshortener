package server

import (
	"bufio"
	"encoding/json"
	"fiber-apis/internal/models"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/url"
	"os"
)

// IsValidURL проверяет, является ли URL допустимым.
func IsValidURL(url1 string) bool {
	_, err := url.ParseRequestURI(url1)
	return err == nil
}

// generateShortID генерирует короткий идентификатор.
func GenerateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY0123456789"
	idLength := 8
	b := make([]byte, idLength)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

// GenerateUserID генерирует идентификатор пользователя.
func GenerateUserID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength := 10
	b := make([]byte, idLength)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

// SaveStorageToFile сохраняет хранилище в файл.
func (s *Server) SaveStorageToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	keys, err := s.Storage.GetAllKeys()
	if err != nil {
		return err
	}

	for _, key := range keys {
		url, err := s.Storage.GetURL(key, "")
		if err != nil {
			return err
		}

		entry := map[string]string{
			"uuid":         key,
			"short_url":    url.ShortURL,
			"original_url": url.OriginalURL,
		}

		entryJSON, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		_, err = writer.WriteString(string(entryJSON) + "\n")
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

// loadStorageFromFile загружает хранилище из файла.
func (s *Server) LoadStorageFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry map[string]string
		err := json.Unmarshal([]byte(scanner.Text()), &entry)
		if err != nil {
			return err
		}

		url := models.URL{
			ShortURL:    entry["short_url"],
			OriginalURL: entry["original_url"],
		}

		_, err = s.Storage.SetURL(url.ShortURL, url.OriginalURL, "")
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// fiberLogger логер от fiber
type fiberLogger struct {
	logger *logrus.Logger
}

// Write записывает данные в лог.
func (f *fiberLogger) Write(p []byte) (n int, err error) {
	f.logger.Info(string(p))
	return len(p), nil
}

func LoadConfigFromFile(filePath string) (*models.Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &models.Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
