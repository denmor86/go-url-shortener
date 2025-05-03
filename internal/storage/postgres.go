package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CheckDSN(dsn string) error {
	_, err := pgxpool.ParseConfig(dsn)
	return err
}

func PingPostrges(ctx context.Context, dsn string, timeout time.Duration) error {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pool, err := pgxpool.New(pingCtx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	return pool.Ping(pingCtx)
}
