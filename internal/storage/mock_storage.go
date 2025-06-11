package storage

import "database/sql"

type MockStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (m *MockStorage) GetDB() *sql.DB {
	return nil
}

func (m *MockStorage) UpdateGauge(name string, value float64) {
	m.Gauges[name] = value
}

func (m *MockStorage) UpdateCounter(name string, value int64) {
	m.Counters[name] += value
}

func (m *MockStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	return m.Gauges, m.Counters
}
