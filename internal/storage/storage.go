package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	"sync"
)

type InternalStorage struct {
	sync.RWMutex
	urls map[string]string
}

func NewInternalStorage() *InternalStorage {
	return &InternalStorage{
		urls: make(map[string]string),
	}
}

func (s *InternalStorage) GetURL(id string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	originalURL, ok := s.urls[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (s *InternalStorage) SetURL(id, url string) {
	s.Lock()
	defer s.Unlock()
	s.urls[id] = url
}

func (s *InternalStorage) GetAllKeys() ([]string, error) {
	s.RLock()
	defer s.RUnlock()
	keys := make([]string, 0, len(s.urls))
	for k := range s.urls {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *InternalStorage) Ping() error {
	return nil // In-memory storage doesn't need to ping a database
}

type DatabaseStorage struct {
	pool *pgxpool.Pool
}

func NewDatabaseStorage(pool *pgxpool.Pool) *DatabaseStorage {
	return &DatabaseStorage{
		pool: pool,
	}
}

func (s *DatabaseStorage) GetURL(id string) (string, error) {
	return "", nil
}

func (s *DatabaseStorage) SetURL(id, url string) {

}

func (s *DatabaseStorage) GetAllKeys() ([]string, error) {
	return nil, nil
}

func (s *DatabaseStorage) Ping() error {
	return s.pool.Ping(context.Background())
}
