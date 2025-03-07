package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"manga-reader2/internal/common/logger"
)

// PostgresConfig содержит настройки подключения к PostgreSQL
type PostgresConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	MaxOpenConn int
	MaxIdleConn int
	MaxLifetime time.Duration
}

// PostgresDB представляет подключение к PostgreSQL
type PostgresDB struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewPostgresDB создает и настраивает новое подключение к PostgreSQL
func NewPostgresDB(ctx context.Context, cfg PostgresConfig, log logger.Logger) (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	log.Info("Подключение к PostgreSQL",
		"host", cfg.Host,
		"port", cfg.Port,
		"dbname", cfg.DBName,
	)

	db, err := sqlx.ConnectContext(ctx, "postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	if cfg.MaxOpenConn > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConn)
	} else {
		db.SetMaxOpenConns(25)
	}

	if cfg.MaxIdleConn > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConn)
	} else {
		db.SetMaxIdleConns(5)
	}

	if cfg.MaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.MaxLifetime)
	} else {
		db.SetConnMaxLifetime(5 * time.Minute)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ошибка проверки соединения с PostgreSQL: %w", err)
	}

	log.Info("Успешное подключение к PostgreSQL")

	return &PostgresDB{
		db:  db,
		log: log,
	}, nil
}

// GetDB возвращает объект базы данных
func (p *PostgresDB) GetDB() *sqlx.DB {
	return p.db
}

// Close закрывает соединение с базой данных
func (p *PostgresDB) Close() error {
	p.log.Info("Закрытие соединения с PostgreSQL")
	return p.db.Close()
}

// Ping проверяет соединение с базой данных
func (p *PostgresDB) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Begin начинает новую транзакцию
func (p *PostgresDB) Begin(ctx context.Context) (*sqlx.Tx, error) {
	return p.db.BeginTxx(ctx, nil)
}

// Exec выполняет SQL-запрос, который не возвращает строки
func (p *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// QueryRow выполняет SQL-запрос, который возвращает одну строку
func (p *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return p.db.QueryRowxContext(ctx, query, args...)
}

// Query выполняет SQL-запрос, который возвращает несколько строк
func (p *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return p.db.QueryxContext(ctx, query, args...)
}

// Get выполняет SQL-запрос и сканирует результат в указанную структуру
func (p *PostgresDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return p.db.GetContext(ctx, dest, query, args...)
}

// Select выполняет SQL-запрос и сканирует результаты в указанный slice
func (p *PostgresDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return p.db.SelectContext(ctx, dest, query, args...)
}

// NamedExec выполняет именованный SQL-запрос
func (p *PostgresDB) NamedExec(ctx context.Context, query string, arg interface{}) (int64, error) {
	result, err := p.db.NamedExecContext(ctx, query, arg)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// NamedQuery выполняет именованный SQL-запрос и возвращает строки
func (p *PostgresDB) NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return p.db.NamedQueryContext(ctx, query, arg)
}
