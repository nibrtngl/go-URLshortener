package localstorage

import (
	"encoding/json"
	"errors"
	"fiber-apis/internal/models"
	"os"
)

// InternalStorage представляет собой структуру, которая представляет простое внутреннее хранилище для URL-адресов.
type InternalStorage struct {
	urls map[string]models.URL
}

// NewInternalStorage создает новый экземпляр InternalStorage.
func NewInternalStorage() *InternalStorage {
	return &InternalStorage{
		urls: make(map[string]models.URL),
	}
}

// Реализация интерфейса Storable

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
			return errors.New("url not found")
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

// Реализация методов для работы с файлами

func (s *InternalStorage) SaveToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(s.urls, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func (s *InternalStorage) LoadFromFile(filePath string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &s.urls)
	return err
}
