package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

func CreateURLsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			short_url TEXT NOT NULL,
			original_url TEXT NOT NULL
		)
	`
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create urls table: %v", err)
	}
	return nil
}

func InitURLsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		INSERT INTO urls (short_url, original_url)
		VALUES ($1, $2)
	`
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create urls table: %v", err)
	}
	return nil
}
