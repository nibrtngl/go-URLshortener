package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func InitDB(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS urls (
            id SERIAL PRIMARY KEY,
            short_url VARCHAR(255) NOT NULL,
            original_url VARCHAR(255) NOT NULL UNIQUE
        )
    `)
	if err != nil {
		return err
	}
	return err
}

type DatabaseStorage struct {
	pool *pgxpool.Pool
}

func NewDatabaseStorage(pool *pgxpool.Pool) *DatabaseStorage {
	return &DatabaseStorage{
		pool: pool,
	}
}

func (s *DatabaseStorage) GetURL(shortURL string) (string, error) {
	query := "SELECT original_url FROM urls WHERE short_url = $1"
	row := s.pool.QueryRow(context.Background(), query, shortURL)

	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("no original URL found for shortURL %s", shortURL)
		}
		return "", fmt.Errorf("failed to get URL from database: %v", err)
	}

	return originalURL, nil
}

func (s *DatabaseStorage) SetURL(id, url string) (string, error) {
	query := "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)"
	_, err := s.pool.Exec(context.Background(), query, id, url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				query := "SELECT short_url FROM urls WHERE original_url = $1"
				row := s.pool.QueryRow(context.Background(), query, url)
				var shortURL string
				if err := row.Scan(&shortURL); err != nil {
					return "", fmt.Errorf("failed to get short URL from database: %v", err)
				}
				return shortURL, nil
			}
		}
		return "", fmt.Errorf("failed to insert URL into database: %v", err)
	}
	return "", nil
}

func (s *DatabaseStorage) GetAllKeys() ([]string, error) {
	return nil, nil
}

func (s *DatabaseStorage) Ping() error {
	return s.pool.Ping(context.Background())
}
