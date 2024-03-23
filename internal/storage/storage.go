package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	"sync"
)

type InMemoryStorage struct {
	sync.RWMutex
	urls map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		urls: make(map[string]string),
	}
}

func (s *InMemoryStorage) GetURL(id string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	originalURL, ok := s.urls[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (s *InMemoryStorage) SetURL(id, url string) {
	s.Lock()
	defer s.Unlock()
	s.urls[id] = url
}

func (s *InMemoryStorage) GetAllKeys() ([]string, error) {
	s.RLock()
	defer s.RUnlock()
	keys := make([]string, 0, len(s.urls))
	for k := range s.urls {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *InMemoryStorage) Ping() error {
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
	// Implement the query to get the URL from the database.
	// Return the URL or an error if the query fails.
	return "", nil // Placeholder
}

func (s *DatabaseStorage) SetURL(id, url string) {
	// Implement the query to store the URL in the database.
}

func (s *DatabaseStorage) GetAllKeys() ([]string, error) {
	// Implement the query to get all keys from the database.
	// Return the keys or an error if the query fails.
	return nil, nil // Placeholder
}

func (s *DatabaseStorage) Ping() error {
	return s.pool.Ping(context.Background())
}
