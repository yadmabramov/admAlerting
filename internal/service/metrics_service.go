package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/yadmabramov/admAlerting/internal/storage"
)

type MetricsService struct {
	storage storage.Repository
	key     string
}

func NewMetricsService(storage storage.Repository, key string) *MetricsService {
	return &MetricsService{
		storage: storage,
		key:     key,
	}
}

func (s *MetricsService) UpdateGauge(ctx context.Context, name string, value string) error {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid gauge value: %w", err)
	}
	return s.storage.UpdateGauge(ctx, name, floatValue)
}

func (s *MetricsService) UpdateCounter(ctx context.Context, name string, value string) error {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value: %w", err)
	}
	return s.storage.UpdateCounter(ctx, name, intValue)
}

func (s *MetricsService) GetGauge(ctx context.Context, name string) (float64, bool) {
	return s.storage.GetGauge(ctx, name)
}

func (s *MetricsService) GetCounter(ctx context.Context, name string) (int64, bool) {
	return s.storage.GetCounter(ctx, name)
}

func (s *MetricsService) GetAllMetrics(ctx context.Context) (map[string]float64, map[string]int64, error) {
	return s.storage.GetAllMetrics(ctx)
}
func (s *MetricsService) GetKey() string {
	return s.key
}
