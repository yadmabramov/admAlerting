package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage(t *testing.T) {
	t.Run("Gauge operations", func(t *testing.T) {
		s := NewMemoryStorage()

		s.UpdateGauge("test_gauge", 123.45)
		s.UpdateGauge("test_gauge", 678.90)

		gauges, _ := s.GetAllMetrics()
		assert.Equal(t, 678.90, gauges["test_gauge"])
	})

	t.Run("Counter operations", func(t *testing.T) {
		s := NewMemoryStorage()

		s.UpdateCounter("test_counter", 10)
		s.UpdateCounter("test_counter", 5)

		_, counters := s.GetAllMetrics()
		assert.Equal(t, int64(15), counters["test_counter"])
	})
}
