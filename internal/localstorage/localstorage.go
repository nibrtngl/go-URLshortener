package localstorage

import (
	"errors"
	"fiber-apis/internal/models"
	"fmt"
)

type InternalStorage struct {
	urls map[string]models.URL
}

func NewInternalStorage() *InternalStorage {
	return &InternalStorage{
		urls: make(map[string]models.URL),
	}
}

func (s *InternalStorage) GetURL(shortURL string, userID string) (models.URL, error) {
	url, ok := s.urls[shortURL]
	if !ok {
		return models.URL{}, errors.New("url not found")
	}
	return url, nil
}

func (s *InternalStorage) SetURL(id, url string, userID string) (string, error) {
	if _, ok := s.urls[id]; ok {
		return "", errors.New("url already exists")
	}
	s.urls[id] = models.URL{
		ShortURL:    id,
		OriginalURL: url,
	}
	return id, nil
}

func (s *InternalStorage) SetURLsAsDeleted(ids []string, userID string) error {
	for _, id := range ids {
		url, ok := s.urls[id]
		if !ok {
			return fmt.Errorf("url not found: %s", id)
		}
		url.IsDeleted = true
		s.urls[id] = url
	}
	return nil
}

func (s *InternalStorage) GetAllKeys() ([]string, error) {
	keys := make([]string, 0, len(s.urls))
	for k := range s.urls {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *InternalStorage) GetUserURLs(userID string) ([]models.URL, error) {
	var urls []models.URL
	for _, url := range s.urls {
		urls = append(urls, url)
	}
	return urls, nil
}

func (s *InternalStorage) Ping() error {
	return nil
}
