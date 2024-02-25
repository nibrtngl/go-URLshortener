package server

import (
	"bufio"
	"encoding/json"
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
		line := scanner.Bytes()
		var entry map[string]string
		err := json.Unmarshal(line, &entry)
		if err != nil {
			logrus.Errorf("Failed to unmarshal storage entry: %v", err)
			continue
		}
		s.Storage[entry["uuid"]] = entry["original_url"]
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

	encoder := json.NewEncoder(file)
	for uuid, originalURL := range s.Storage {
		entry := map[string]string{
			"uuid":         uuid,
			"original_url": originalURL,
		}
		err := encoder.Encode(entry)
		if err != nil {
			logrus.Errorf("Failed to encode storage entry: %v", err)
			continue
		}
	}

	return nil
}
