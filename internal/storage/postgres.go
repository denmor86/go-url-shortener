package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UniqueViolationError struct {
	Message  string
	ShortURL string
}

func (e *UniqueViolationError) Error() string {
	return e.Message
}

type DatabaseStorage struct {
	Pool   *pgxpool.Pool
	Config *pgx.ConnConfig
}

const (
	CheckExist     = `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname =$1)`
	CreateDatabase = `CREATE DATABASE %s`
	CreateTable    = `CREATE TABLE IF NOT EXISTS URLs (
	                       id SERIAL PRIMARY KEY,
	                       short_url TEXT  NOT NUll,
	                       original_url TEXT UNIQUE NOT NUll
						);`
	InsertRecord = `INSERT INTO URLs (short_url, original_url) 
						VALUES ($1, $2) 
						ON CONFLICT (original_url) DO NOTHING
						RETURNING short_url;`
	GetOriginalURL = `SELECT original_url FROM URLs WHERE short_url =$1;`
	GetShortURL    = `SELECT short_url FROM URLs WHERE original_url =$1;`
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
	return &DatabaseStorage{Pool: pool, Config: cfg.ConnConfig}, nil
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
	s.Pool.Close()
	return nil
}

func (s *DatabaseStorage) CreateDatabase(ctx context.Context) error {
	conn, err := pgx.ConnectConfig(ctx, s.Config)
	if err != nil {
		// если не получилось соединиться с БД из строки подключения
		// пробуем использовать дефолтную БД
		cfg := s.Config.Copy()
		cfg.Database = `postgres`
		conn, err = pgx.ConnectConfig(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to connect database: %w", err)
		}
	}
	defer conn.Close(ctx)

	var exist bool
	err = conn.QueryRow(ctx, CheckExist, s.Config.Database).Scan(&exist)
	if err != nil {
		return fmt.Errorf("failed to check database exists: %w", err)
	}
	if !exist {
		_, err = conn.Exec(ctx, fmt.Sprintf(CreateDatabase, s.Config.Database))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}
	return nil
}

func (s *DatabaseStorage) CreateTable(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, CreateTable)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (s *DatabaseStorage) Add(ctx context.Context, originalURL string, shortURL string) error {

	var prevShortURL string
	err := s.Pool.QueryRow(ctx, InsertRecord, shortURL, originalURL).Scan(&prevShortURL)
	// добавили в базу, совпадений нет
	if err == nil {
		return nil
	}
	// ошибка добавления строки
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to add record: %w", err)
	}
	// есть совпадение оригинального адреса
	err = s.Pool.QueryRow(ctx, GetShortURL, originalURL).Scan(&prevShortURL)
	if err != nil {
		return fmt.Errorf("failed to get record: %w", err)
	}
	return &UniqueViolationError{Message: "URL already exists", ShortURL: prevShortURL}
}

func (s *DatabaseStorage) AddMultiple(ctx context.Context, items []TableItem) error {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	defer func() {
		err = tx.Rollback(ctx)
	}()

	for _, url := range items {
		_, err := s.Pool.Exec(ctx, InsertRecord, url.ShortURL, url.OriginalURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *DatabaseStorage) Get(ctx context.Context, shortURL string) (string, error) {

	var originalURL string
	err := s.Pool.QueryRow(ctx, GetOriginalURL, shortURL).Scan(&originalURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("short url not found: %s", shortURL)
		}
		return "", fmt.Errorf("failed to get record: %w", err)
	}
	return originalURL, nil
}

func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.Pool.Ping(ctx)
}
