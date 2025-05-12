package service

import (
	"fmt"
	"strconv"

	"github.com/yadmabramov/admAlerting/internal/storage"
)

type MetricsService struct {
	storage storage.Repository
}

func NewMetricsService(storage storage.Repository) *MetricsService {
	return &MetricsService{storage: storage}
}

func (s *MetricsService) UpdateGauge(name string, value string) error {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid gauge value: %w", err)
	}
	return s.storage.UpdateGauge(name, floatValue)
}

func (s *MetricsService) UpdateCounter(name string, value string) error {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value: %w", err)
	}
	return s.storage.UpdateCounter(name, intValue)
}

func (s *MetricsService) GetGauge(name string) (float64, bool) {
	return s.storage.GetGauge(name)
}

func (s *MetricsService) GetCounter(name string) (int64, bool) {
	return s.storage.GetCounter(name)
}

func (s *MetricsService) GetAllMetrics() (map[string]float64, map[string]int64) {
	return s.storage.GetAllMetrics()
}
