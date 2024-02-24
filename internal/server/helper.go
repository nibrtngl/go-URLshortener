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

func (s *Server) loadStorageFromFile(filePath string) error {
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
		s.Storage[entry["short_url"]] = entry["original_url"]
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (s *Server) saveStorageToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for shortURL, originalURL := range s.Storage {
		entry := map[string]string{
			"short_url":    shortURL,
			"original_url": originalURL,
		}
		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		_, err = file.Write(data)
		if err != nil {
			return err
		}
		_, err = file.WriteString("\n")
		if err != nil {
			return err
		}
	}

	return nil
}
