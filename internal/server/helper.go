package server

import (
	"bufio"
	"encoding/json"
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

	// Generate a unique identifier
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func (s *models.Server) saveStorageToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for key, value := range s.Storage {
		entry := map[string]string{
			"uuid":         key,
			"short_url":    value,
			"original_url": value,
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

func (s *models.Server) loadStorageFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry map[string]string
		err := json.Unmarshal(scanner.Bytes(), &entry)
		if err != nil {
			return err
		}

		uuid := entry["uuid"]
		shortURL := entry["short_url"]
		originalURL := entry["original_url"]

		s.Storage[shortURL] = originalURL
		s.Storage[uuid] = shortURL
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
