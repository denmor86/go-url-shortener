package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseStorage struct {
	pool   *pgxpool.Pool
	config *pgx.ConnConfig
	dbHOST string
}

const (
	CheckExistSQL     = `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname =$1)`
	CreateDatabaseSQL = `CREATE DATABASE %s`
	CreateTableSQL    = `CREATE TABLE IF NOT EXISTS URLs (
	                       id SERIAL PRIMARY KEY,
	                       short_url TEXT NOT NUll,
	                       base_url TEXT NOT NUll
						);`
	InsertRecordSQL = `INSERT INTO URLs (short_url, base_url) VALUES ($1, $2);`
	GetRecordSQL    = `SELECT base_url FROM URLs WHERE short_url =$1;`
)

func NewDatabaseStorage(dsn string) (*DatabaseStorage, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}
	return &DatabaseStorage{pool: pool, config: cfg.ConnConfig}, nil
}
func (s *DatabaseStorage) Initialize() error {
	if err := s.CreateDatabase(context.Background()); err != nil {
		return fmt.Errorf("error create database: %w", err)
	}
	if err := s.CreateTable(context.Background()); err != nil {
		return fmt.Errorf("error create tables: %w", err)
	}
	return nil
}

func (s *DatabaseStorage) Close() error {
	s.pool.Close()
	return nil
}

func (s *DatabaseStorage) CreateDatabase(ctx context.Context) error {
	conn, err := pgx.ConnectConfig(ctx, s.config)
	if err != nil {
		// если не получилось соединиться с БД из строки подключения
		// пробуем использовать дефолтную БД
		cfg := s.config.Copy()
		cfg.Database = `postgres`
		conn, err = pgx.ConnectConfig(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to connect database: %w", err)
		}
	}
	defer conn.Close(ctx)

	var exist bool
	err = conn.QueryRow(ctx, CheckExistSQL, s.config.Database).Scan(&exist)
	if err != nil {
		return fmt.Errorf("failed to check database exists: %w", err)
	}
	if !exist {
		_, err = conn.Exec(ctx, fmt.Sprintf(CreateDatabaseSQL, s.config.Database))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}
	return nil
}

func (s *DatabaseStorage) CreateTable(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, CreateTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (s *DatabaseStorage) Add(ctx context.Context, longURL string, shortURL string) error {

	_, err := s.pool.Exec(ctx, InsertRecordSQL, shortURL, longURL)
	if err != nil {
		return fmt.Errorf("failed to add record: %w", err)
	}

	return nil
}

func (s *DatabaseStorage) Get(ctx context.Context, shortURL string) (string, error) {

	var baseURL string
	err := s.pool.QueryRow(ctx, GetRecordSQL, shortURL).Scan(&baseURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("short url not found: %s", shortURL)
		}
		return "", fmt.Errorf("failed to get record: %w", err)
	}
	return baseURL, nil
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
