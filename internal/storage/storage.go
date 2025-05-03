package storage

type Repository interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetAllMetrics() (gauges map[string]float64, counters map[string]int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}
