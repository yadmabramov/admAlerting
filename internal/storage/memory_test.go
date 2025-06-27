package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStorage()

	t.Run("Gauge operations", func(t *testing.T) {
		err := s.UpdateGauge(ctx, "test_gauge", 123.45)
		assert.NoError(t, err)
		err = s.UpdateGauge(ctx, "test_gauge", 678.90)
		assert.NoError(t, err)

		gauges, _, err := s.GetAllMetrics(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 678.90, gauges["test_gauge"])
	})

	t.Run("Counter operations", func(t *testing.T) {
		err := s.UpdateCounter(ctx, "test_counter", 10)
		assert.NoError(t, err)
		err = s.UpdateCounter(ctx, "test_counter", 5)
		assert.NoError(t, err)

		_, counters, err := s.GetAllMetrics(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), counters["test_counter"])
	})
}
