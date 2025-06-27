package storage

import (
	"context"
	"database/sql"
)

type Repository interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	GetAllMetrics(ctx context.Context) (gauges map[string]float64, counters map[string]int64, err error)
	GetGauge(ctx context.Context, name string) (float64, bool)
	GetCounter(ctx context.Context, name string) (int64, bool)
	GetDB() *sql.DB
	Close() error
}
