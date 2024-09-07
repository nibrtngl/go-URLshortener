package db

import (
	"context"
	"errors"
	"fiber-apis/internal/models"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// InitDB инициализирует базу данных, создавая необходимые таблицы.
func InitDB(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS urls (
            short_url VARCHAR(255),
            original_url VARCHAR(255) PRIMARY KEY,
            is_deleted BOOLEAN DEFAULT FALSE,
            user_id VARCHAR(255)
        )
    `)
	if err != nil {
		return err
	}
	return err
}

// SaveToFile не используется в базе данных, поэтому возвращает nil.
func (s *DatabaseStorage) SaveToFile(path string) error {
	// В случае работы с БД, обычно данные хранятся напрямую в БД, так что сохранять их в файл не требуется.
	return nil
}

// LoadFromFile не используется в базе данных, поэтому возвращает nil.
func (s *DatabaseStorage) LoadFromFile(path string) error {
	// Загрузка данных из файла не требуется для БД.
	return nil
}

// DatabaseStorage представляет собой структуру, которая представляет хранилище базы данных.
type DatabaseStorage struct {
	pool *pgxpool.Pool
}

// NewDatabaseStorage создает новый экземпляр DatabaseStorage.
func NewDatabaseStorage(pool *pgxpool.Pool) *DatabaseStorage {
	return &DatabaseStorage{
		pool: pool,
	}
}

// GetURL извлекает URL из базы данных на основе предоставленного shortURL и userID.
func (s *DatabaseStorage) GetURL(shortURL string, userID string) (models.URL, error) {
	query := "SELECT original_url, is_deleted FROM urls WHERE short_url = $1 AND user_id = $2"
	row := s.pool.QueryRow(context.Background(), query, shortURL, userID)

	var originalURL string
	var isDeleted bool
	err := row.Scan(&originalURL, &isDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.URL{}, fmt.Errorf("no original URL found for shortURL %s", shortURL)
		}
		return models.URL{}, fmt.Errorf("failed to get URL from database: %v", err)
	}

	return models.URL{ShortURL: shortURL, OriginalURL: originalURL, IsDeleted: isDeleted}, nil
}

// SetURL добавляет новый URL в базу данных и возвращает сгенерированный ID.
func (s *DatabaseStorage) SetURL(id, url string, userID string) (string, error) {
	query := `
        INSERT INTO urls (short_url, original_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (original_url) DO NOTHING
    `
	_, err := s.pool.Exec(context.Background(), query, id, url, userID)
	if err != nil {
		return "", fmt.Errorf("failed to insert or retrieve URL from database: %v", err)
	}
	query = `
     SELECT short_url FROM urls WHERE original_url = $1 AND user_id = $2;`
	row := s.pool.QueryRow(context.Background(), query, url, userID)
	var shortURL string
	err = row.Scan(&shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to insert or retrieve URL from database: %v", err)
	}
	return shortURL, nil
}

// SetURLsAsDeleted помечает предоставленные URL-адреса как удаленные в базе данных.
func (s *DatabaseStorage) SetURLsAsDeleted(ids []string, userID string) error {
	query := `
        UPDATE urls
        SET is_deleted = true
        WHERE short_url = ANY($1) AND user_id = $2
    `
	_, err := s.pool.Exec(context.Background(), query, ids, userID)
	if err != nil {
		return fmt.Errorf("failed to set URLs as deleted: %v", err)
	}
	return nil
}

// GetAllKeys извлекает все ключи из базы данных.
func (s *DatabaseStorage) GetAllKeys() ([]string, error) {
	return nil, nil
}

// GetUserURLs извлекает все URL-адреса, связанные с предоставленным userID, из базы данных.
func (s *DatabaseStorage) GetUserURLs(userID string) ([]models.URL, error) {
	query := "SELECT short_url, original_url FROM urls WHERE user_id = $1"
	rows, err := s.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs from database: %v", err)
	}
	defer rows.Close()

	var urls []models.URL
	for rows.Next() {
		var url models.URL
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan user URL: %v", err)
		}
		urls = append(urls, url)
	}

	return urls, nil
}

// Ping проверяет состояние базы данных.
func (s *DatabaseStorage) Ping() error {
	return s.pool.Ping(context.Background())
}
