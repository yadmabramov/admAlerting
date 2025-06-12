package storage

import "database/sql"

type Repository interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetAllMetrics() (gauges map[string]float64, counters map[string]int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetDB() *sql.DB
	Close() error
}
