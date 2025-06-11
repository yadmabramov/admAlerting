package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	DB *sql.DB
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PostgresStorage{DB: db}, nil
}

func createTables(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS gauges (
			name VARCHAR(255) PRIMARY KEY,
			value DOUBLE PRECISION NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create gauges table: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS counters (
			name VARCHAR(255) PRIMARY KEY,
			value BIGINT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create counters table: %w", err)
	}

	return nil
}

func (s *PostgresStorage) UpdateGauge(name string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO gauges (name, value) 
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, name, value)

	return err
}

func (s *PostgresStorage) UpdateCounter(name string, value int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO counters (name, value) 
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET value = counters.value + EXCLUDED.value, updated_at = NOW()
	`, name, value)

	return err
}

func (s *PostgresStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, "SELECT name, value FROM gauges")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var value float64
			if err := rows.Scan(&name, &value); err == nil {
				gauges[name] = value
			}
		}
		if err := rows.Err(); err != nil {
			log.Printf("Error reading gauges: %v", err)
		}
	}

	rows, err = s.DB.QueryContext(ctx, "SELECT name, value FROM counters")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var value int64
			if err := rows.Scan(&name, &value); err == nil {
				counters[name] = value
			}
		}
		if err := rows.Err(); err != nil {
			log.Printf("Error reading counters: %v", err)
		}
	}

	return gauges, counters
}

func (s *PostgresStorage) GetGauge(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var value float64
	err := s.DB.QueryRowContext(ctx, "SELECT value FROM gauges WHERE name = $1", name).Scan(&value)
	if err != nil {
		return 0, false
	}
	return value, true
}

func (s *PostgresStorage) GetCounter(name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var value int64
	err := s.DB.QueryRowContext(ctx, "SELECT value FROM counters WHERE name = $1", name).Scan(&value)
	if err != nil {
		return 0, false
	}
	return value, true
}

func (s *PostgresStorage) Close() error {
	return s.DB.Close()
}

func (s *PostgresStorage) GetDB() *sql.DB {
	return s.DB
}
