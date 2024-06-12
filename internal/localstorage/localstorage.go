package localstorage

import (
	"errors"
)

type InternalStorage struct {
	urls map[string]string
}

func NewInternalStorage() *InternalStorage {
	return &InternalStorage{
		urls: make(map[string]string),
	}
}

func (s *InternalStorage) GetURL(shortURL string, userID string) (string, error) {
	originalURL, ok := s.urls[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return originalURL, nil
}

func (s *InternalStorage) SetURL(id, url string, userID string) (string, error) {
	if _, ok := s.urls[id]; ok {
		return "", errors.New("url already exists")
	}
	s.urls[id] = url
	return id, nil
}

func (s *InternalStorage) GetAllKeys() ([]string, error) {
	keys := make([]string, 0, len(s.urls))
	for k := range s.urls {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *InternalStorage) Ping() error {
	return nil
}
