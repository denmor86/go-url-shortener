package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// UniqueViolation - ошибка нарушения уникальности URL
type UniqueViolation struct {
	Message  string // сообщение с ошибкой
	ShortURL string // короткая ссылка
}

// Error - метод получения текста ошибки
func (e *UniqueViolation) Error() string {
	return e.Message
}

// DeletedViolation - ошибка нарушения ранее удаленного URL
type DeletedViolation struct {
	Message string
}

// Error - метод получения текста ошибки
func (e *DeletedViolation) Error() string {
	return e.Message
}

// DatabaseStorage - хранилище данных в БД
type DatabaseStorage struct {
	Pool   *pgxpool.Pool   // пул подключений
	Config *pgx.ConnConfig // конфигурация БД
	DSN    string          // строка подключения БД
}

// Используемые SQL запросы
const (
	// CheckExist - SQL запрос для проверки наличи БД
	CheckExist = `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname =$1)`
	// CreateDatabase - SQL запрос для создания БД
	CreateDatabase = `CREATE DATABASE %s`
	// InsertRecord - SQL запрос добавления записи по URL
	InsertRecord = `INSERT INTO URLs (short_url, original_url, user_uuid) 
						VALUES ($1, $2, $3) 
						ON CONFLICT (original_url) DO NOTHING
						RETURNING short_url;`
	// GetOriginalURL - SQL запрос получения оригинальной URL по короткой записи
	GetOriginalURL = `SELECT original_url, is_deleted FROM URLs WHERE short_url =$1;`
	// GetShortURL - SQL запрос получения короткой записи по оригинаольному URL
	GetShortURL = `SELECT short_url FROM URLs WHERE original_url =$1;`
	// GetUserlURL - SQL запрос записи по пользователю
	GetUserlURL = `SELECT user_uuid, original_url, short_url FROM urls WHERE user_uuid=$1 AND NOT is_deleted;`
	// DeleteUserURL - SQL запрос отметки записи для удаления по пользователю и короткой ссылке
	DeleteUserURL = `UPDATE urls SET is_deleted=TRUE WHERE user_uuid=$1 AND short_url=$2`
)

// NewDatabaseStorage - метод создания хранилища данных в БД
func NewDatabaseStorage(dsn string) (*DatabaseStorage, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}
	return &DatabaseStorage{Pool: pool, Config: cfg.ConnConfig, DSN: dsn}, nil
}

// Initialize - метод инициализации хранилища(создание БД, миграции)
func (s *DatabaseStorage) Initialize() error {

	if err := s.CreateDatabase(context.Background()); err != nil {
		return fmt.Errorf("error create database: %w", err)
	}
	if err := Migration(s.DSN); err != nil {
		return fmt.Errorf("error migrate database: %w", err)
	}

	return nil
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Migration - метод миграций таблиц БД
func Migration(DatabaseDSN string) error {

	db, err := sql.Open("pgx", DatabaseDSN)
	if err != nil {
		return fmt.Errorf("open db error: %w ", err)
	}
	defer db.Close()
	// используется для внутренней файловой системы (загруженные ресурсы)
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect error: %w ", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose run migrations error:  %w ", err)
	}
	return nil
}

// Close - метод закрытия соединения с БД
func (s *DatabaseStorage) Close() error {
	s.Pool.Close()
	return nil
}

// CreateDatabase - метод создания БД
func (s *DatabaseStorage) CreateDatabase(ctx context.Context) error {
	// goose не умеет создавать БД
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
	}
	defer conn.Close(ctx)
	return nil
}

// AddRecord - метод добавления записи в БД
func (s *DatabaseStorage) AddRecord(ctx context.Context, record TableRecord) error {

	var prevShortURL string
	err := s.Pool.QueryRow(ctx, InsertRecord, record.ShortURL, record.OriginalURL, record.UserID).Scan(&prevShortURL)
	// добавили в базу, совпадений нет
	if err == nil {
		return nil
	}
	// ошибка добавления строки
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to add record: %w", err)
	}
	// есть совпадение оригинального адреса
	err = s.Pool.QueryRow(ctx, GetShortURL, record.OriginalURL).Scan(&prevShortURL)
	if err != nil {
		return fmt.Errorf("failed to get record: %w", err)
	}
	return &UniqueViolation{Message: "URL already exists", ShortURL: prevShortURL}
}

// AddRecords - метод добавления массива записей в БД
func (s *DatabaseStorage) AddRecords(ctx context.Context, records []TableRecord) error {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	defer func() {
		err = tx.Rollback(ctx)
	}()

	for _, rec := range records {
		_, err := s.Pool.Exec(ctx, InsertRecord, rec.ShortURL, rec.OriginalURL, rec.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetRecord - метод получения записи по короткой ссылке
func (s *DatabaseStorage) GetRecord(ctx context.Context, shortURL string) (string, error) {

	var originalURL string
	var isDeleted bool
	err := s.Pool.QueryRow(ctx, GetOriginalURL, shortURL).Scan(&originalURL, &isDeleted)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("short url not found: %s", shortURL)
		}
		return "", fmt.Errorf("failed to get record: %w", err)
	}
	if isDeleted {
		return originalURL, &DeletedViolation{Message: "URL is deleted"}
	}
	return originalURL, nil
}

// GetUserRecords - метод получения массива записей пользователя из БД
func (s *DatabaseStorage) GetUserRecords(ctx context.Context, userID string) ([]TableRecord, error) {
	var records []TableRecord

	rows, err := s.Pool.Query(ctx, GetUserlURL, userID)
	if err != nil {
		return records, fmt.Errorf("failed to get user record: %w", err)
	}
	for rows.Next() {
		var record TableRecord
		err = rows.Scan(&record.UserID, &record.OriginalURL, &record.ShortURL)
		if err != nil {
			return records, fmt.Errorf("failed scan  user record: %w", err)
		}
		records = append(records, record)
	}
	return records, nil
}

// DeleteURLs - метод отметки массива записей пользователя на удаление
func (s *DatabaseStorage) DeleteURLs(ctx context.Context, userID string, shortURLS []string) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	batch := &pgx.Batch{}
	for _, rec := range shortURLS {
		batch.Queue(DeleteUserURL, userID, rec)
	}
	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	_, err = br.Exec()
	if err != nil {
		return fmt.Errorf("error executing batch: %w", err)
	}

	err = br.Close()
	if err != nil {
		return fmt.Errorf("error closing batch: %w", err)
	}

	return tx.Commit(ctx)
}

// Ping - метод проверки соединения с БД
func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.Pool.Ping(ctx)
}
