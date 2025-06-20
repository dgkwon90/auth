// Package database provides database connection and utility functions.
package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

var pool *pgxpool.Pool

// Connect initializes the pgxpool connection
func Connect(dataSourceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dataSourceName)
	if err != nil {
		slog.Error("pgxpool config parse error", "error", err)
		return err
	}
	// 풀 사이즈 등 커스터마이즈 가능
	cfg.MaxConns = 10

	pool, err = pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		slog.Error("Error connecting to database (pgxpool)", "error", err)
		return err
	}
	slog.Info("pgxpool connection established")
	return nil
}

// GetPool returns the pgxpool
func GetPool() *pgxpool.Pool {
	return pool
}
