package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
)

var db *pgx.Conn

func ConnectToDB(dsn string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	// Проверяем соединение с базой данных
	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return conn, nil
}
