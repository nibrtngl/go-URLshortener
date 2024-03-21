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

func (s *InMemoryStorage) Ping() error {
	// Since InMemoryStorage doesn't have a connection to a database,
	// we can just return nil to indicate that the storage is available.
	return nil
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

type DatabaseStorage struct {
	pool *pgxpool.Pool
}

func NewDatabaseStorage(pool *pgxpool.Pool) *DatabaseStorage {
	return &DatabaseStorage{
		pool: pool,
	}
}

func (s *DatabaseStorage) GetAllKeys() ([]string, error) {
	rows, err := s.pool.Query(context.Background(), "SELECT id FROM urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		err := rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (s *DatabaseStorage) Ping() error {
	return s.pool.Ping(context.Background())
}
