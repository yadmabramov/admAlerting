package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type MemStorage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

type MemStorageImpl struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorageImpl {
	return &MemStorageImpl{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorageImpl) UpdateGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *MemStorageImpl) UpdateCounter(name string, value int64) {
	s.counters[name] += value
}

func (s *MemStorageImpl) GetAllMetrics() map[string]interface{} {

	result := make(map[string]interface{})
	for k, v := range s.gauges {
		result[k] = v
	}
	for k, v := range s.counters {
		result[k] = v
	}
	return result
}

func main() {
	storage := NewMemStorage()

	http.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
		if len(parts) != 3 {
			http.Error(w, "Invalid URL format", http.StatusNotFound)
			return
		}

		metricType := parts[0]
		metricName := parts[1]
		metricValue := parts[2]

		if metricName == "" {
			http.Error(w, "Metric name cannot be empty", http.StatusNotFound)
			return
		}

		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(metricName, value)
			w.WriteHeader(http.StatusOK)

		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(metricName, value)
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
	})

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
