package storage

import "sync"

type MemoryStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemoryStorage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
}

func (s *MemoryStorage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value
}

func (s *MemoryStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gaugesCopy := make(map[string]float64)
	countersCopy := make(map[string]int64)

	for k, v := range s.gauges {
		gaugesCopy[k] = v
	}

	for k, v := range s.counters {
		countersCopy[k] = v
	}

	return gaugesCopy, countersCopy
}

func (s *MemoryStorage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.gauges[name]
	return val, ok
}

func (s *MemoryStorage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.counters[name]
	return val, ok
}
