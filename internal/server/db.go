package server

import (
	"context"
	"github.com/jackc/pgx/v4"
)

func (s *Server) ConnectDB() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), s.DSN)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Server) PingDB() error {
	conn, err := pgx.Connect(context.Background(), s.DSN)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	var result int
	err = conn.QueryRow(context.Background(), "SELECT 1").Scan(&result)
	if err != nil {
		return err
	}
	return nil
}
