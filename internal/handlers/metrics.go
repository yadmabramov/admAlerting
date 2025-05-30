package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yadmabramov/admAlerting/internal/models"
	"github.com/yadmabramov/admAlerting/internal/service"
)

type MetricsHandler struct {
	service *service.MetricsService
}

func NewMetricsHandler(service *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{service: service}
}

func (h *MetricsHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")
	mValue := chi.URLParam(r, "value")

	var err error
	switch mType {
	case "gauge":
		err = h.service.UpdateGauge(mName, mValue)
	case "counter":
		err = h.service.UpdateCounter(mName, mValue)
	default:
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) HandleGetMetric(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	switch mType {
	case "gauge":
		if value, ok := h.service.GetGauge(mName); ok {
			w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
			return
		}
	case "counter":
		if value, ok := h.service.GetCounter(mName); ok {
			w.Write([]byte(strconv.FormatInt(value, 10)))
			return
		}
	default:
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	http.Error(w, "Metric not found", http.StatusNotFound)
}

func (h *MetricsHandler) HandleGetAllMetricsJSON(w http.ResponseWriter, r *http.Request) {
	gauges, counters := h.service.GetAllMetrics()

	type MetricsResponse struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	response := MetricsResponse{
		Gauges:   gauges,
		Counters: counters,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *MetricsHandler) HandleUpdateJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var response models.Metrics
	var err error

	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			http.Error(w, "Value is required for gauge", http.StatusBadRequest)
			return
		}
		err = h.service.UpdateGauge(metric.ID, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
		if err == nil {
			val, _ := h.service.GetGauge(metric.ID)
			response = models.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Value: &val,
			}
		}
	case "counter":
		if metric.Delta == nil {
			http.Error(w, "Delta is required for counter", http.StatusBadRequest)
			return
		}
		err = h.service.UpdateCounter(metric.ID, strconv.FormatInt(*metric.Delta, 10))
		if err == nil {
			val, _ := h.service.GetCounter(metric.ID)
			response = models.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: &val,
			}
		}
	default:
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *MetricsHandler) HandleGetMetricJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var response models.Metrics

	switch metric.MType {
	case "gauge":
		if value, ok := h.service.GetGauge(metric.ID); ok {
			response = models.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Value: &value,
			}
		} else {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
	case "counter":
		if value, ok := h.service.GetCounter(metric.ID); ok {
			response = models.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: &value,
			}
		} else {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
	default:
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
