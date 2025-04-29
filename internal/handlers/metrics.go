package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid path", http.StatusNotFound)
		return
	}

	mType, mName, mValue := parts[0], parts[1], parts[2]

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

func (h *MetricsHandler) HandleGetAllMetrics(w http.ResponseWriter, r *http.Request) {
	gauges, counters := h.storage.(*storage.MemoryStorage).GetAllMetrics()

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
