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

func (s *InternalStorage) GetURL(shortURL string) (string, error) {
	originalURL, ok := s.urls[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return originalURL, nil
}

func (s *InternalStorage) SetURL(id, url string) (string, error) {
	s.urls[id] = url
	return url, nil
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
