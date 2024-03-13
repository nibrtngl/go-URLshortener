package db

import (
	"fiber-apis/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

func GetDBConnectionParams(cfg *models.Config) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	return poolConfig, nil
}
