package server

import (
	"errors"
	"sync"
)

type MyStorage struct {
	sync.RWMutex
	data map[string]string
}

func (s *MyStorage) GetURL(id string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	originalURL, ok := s.data[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (s *MyStorage) SetURL(id, url string) {
	s.Lock()
	defer s.Unlock()
	s.data[id] = url
}

func (s *MyStorage) GetAllKeys() ([]string, error) {
	s.RLock()
	defer s.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *MyStorage) Ping() error {

	return nil
}

// InitDB is a placeholder method to satisfy the Storable interface.
// In a real-world scenario, it would initialize the database connection.
func (s *MyStorage) InitDB() error {
	// For demonstration purposes, we'll just return nil.
	// In a real implementation, you would initialize the database connection.
	return nil
}
