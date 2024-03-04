package db

import (
	"database/sql"
)

type Database struct {
	Conn *sql.DB
}

func (b *Database) Ping() error {
	return b.Conn.Ping()
}
