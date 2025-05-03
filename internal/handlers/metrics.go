package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yadmabramov/admAlerting/internal/storage"
)

type MetricsHandler struct {
	storage storage.Repository
}

func NewMetricsHandler(storage storage.Repository) *MetricsHandler {
	return &MetricsHandler{storage: storage}
}

func (h *MetricsHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")
	mValue := chi.URLParam(r, "value")

	switch mType {
	case "gauge":
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			http.Error(w, "Invalid value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(mName, value)

	case "counter":
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(mName, value)

	default:
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) HandleGetMetric(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	switch mType {
	case "gauge":
		if value, ok := h.storage.(*storage.MemoryStorage).GetGauge(mName); ok {
			w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
			return
		}
	case "counter":
		if value, ok := h.storage.(*storage.MemoryStorage).GetCounter(mName); ok {
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
	gauges, counters := h.storage.GetAllMetrics()

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
