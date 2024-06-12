package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func InitDB(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS urls (
            short_url VARCHAR(255) ,
            original_url VARCHAR(255) PRIMARY KEY 
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

func (s *DatabaseStorage) GetURL(shortURL string, userID string) (string, error) {
	query := "SELECT original_url FROM urls WHERE short_url = $1"
	row := s.pool.QueryRow(context.Background(), query, shortURL, userID)

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
func (s *DatabaseStorage) GetUserURLs(userID string) ([]string, error) {
	// метод для получения всех URL пользователя
	query := "SELECT original_url FROM urls WHERE user_id = $1"
	rows, err := s.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve URLs from database: %v", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %v", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to retrieve URLs from database: %v", err)
	}

	return urls, nil
}

func (s *DatabaseStorage) SetURL(id, url string, userID string) (string, error) {
	query := `
        INSERT INTO urls (short_url, original_url) 
        VALUES ($1, $2) 
        ON CONFLICT (original_url) DO NOTHING
    `
	_, err := s.pool.Exec(context.Background(), query, id, url, userID)
	if err != nil {
		return "", fmt.Errorf("failed to insert or retrieve URL from database: %v", err)
	}
	query = `
     SELECT short_url FROM urls WHERE original_url = $1;`
	row := s.pool.QueryRow(context.Background(), query, url)
	var shortURL string
	err = row.Scan(&shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to insert or retrieve URL from database: %v", err)
	}
	return shortURL, nil
}

func (s *DatabaseStorage) GetAllKeys() ([]string, error) {
	return nil, nil
}

func (s *DatabaseStorage) Ping() error {
	return s.pool.Ping(context.Background())
}
