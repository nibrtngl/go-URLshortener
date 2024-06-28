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

func isValidURL(url1 string) bool {
	_, err := url.ParseRequestURI(url1)
	return err == nil
}

func generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY0123456789"
	idLength := 8
	b := make([]byte, idLength)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func generateUserID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength := 10
	b := make([]byte, idLength)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func (s *Server) saveStorageToFile(filePath string) error {
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

func (s *Server) loadStorageFromFile(filePath string) error {
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

type fiberLogger struct {
	logger *logrus.Logger
}

func (f *fiberLogger) Write(p []byte) (n int, err error) {
	f.logger.Info(string(p))
	return len(p), nil
}
