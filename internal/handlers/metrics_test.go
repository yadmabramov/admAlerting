package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/yadmabramov/admAlerting/internal/models"
	"github.com/yadmabramov/admAlerting/internal/service"
)

type MockStorage struct {
	lastGauge   float64
	lastCounter int64
}

func (m *MockStorage) UpdateGauge(name string, value float64) error {
	m.lastGauge = value
	return nil
}

func (m *MockStorage) UpdateCounter(name string, value int64) error {
	m.lastCounter = value
	return nil
}

func (m *MockStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	return nil, nil
}

func (m *MockStorage) GetGauge(name string) (float64, bool) {
	return m.lastGauge, true
}

func (m *MockStorage) GetCounter(name string) (int64, bool) {
	return m.lastCounter, true
}

func TestMetricsHandler(t *testing.T) {
	mockStorage := &MockStorage{}
	service := service.NewMetricsService(mockStorage)
	handler := NewMetricsHandler(service)

	t.Run("Update gauge", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/update/gauge/test/123.45", nil)

		// Создаем роутер и добавляем обработчик
		router := http.NewServeMux()
		router.HandleFunc("/update/gauge/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
			handler.HandleUpdate(w, r)
		})

		// Используем chi для параметров URL
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("type", "gauge")
		chiCtx.URLParams.Add("name", "test")
		chiCtx.URLParams.Add("value", "123.45")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

		handler.HandleUpdate(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 123.45, mockStorage.lastGauge)
	})

	t.Run("Update counter", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/update/counter/test/10", nil)

		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("type", "counter")
		chiCtx.URLParams.Add("name", "test")
		chiCtx.URLParams.Add("value", "10")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

		handler.HandleUpdate(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, int64(10), mockStorage.lastCounter)
	})

	t.Run("Invalid method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/update/gauge/test/123.45", nil)

		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("type", "gauge")
		chiCtx.URLParams.Add("name", "test")
		chiCtx.URLParams.Add("value", "123.45")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

		handler.HandleUpdate(w, r)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestMetricsHandlerJSON(t *testing.T) {
	mockStorage := &MockStorage{}
	service := service.NewMetricsService(mockStorage)
	handler := NewMetricsHandler(service)

	t.Run("Update gauge via JSON", func(t *testing.T) {
		metric := models.Metrics{
			ID:    "test_gauge",
			MType: "gauge",
			Value: ptrFloat64(123.45),
		}

		body, _ := json.Marshal(metric)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		handler.HandleUpdateJSON(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 123.45, mockStorage.lastGauge)

		var response models.Metrics
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, metric.ID, response.ID)
		assert.Equal(t, metric.MType, response.MType)
		assert.Equal(t, *metric.Value, *response.Value)
	})

	t.Run("Update counter via JSON", func(t *testing.T) {
		metric := models.Metrics{
			ID:    "test_counter",
			MType: "counter",
			Delta: ptrInt64(10),
		}

		body, _ := json.Marshal(metric)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		handler.HandleUpdateJSON(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, int64(10), mockStorage.lastCounter)

		var response models.Metrics
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, metric.ID, response.ID)
		assert.Equal(t, metric.MType, response.MType)
		assert.Equal(t, *metric.Delta, *response.Delta)
	})

	t.Run("Get metric via JSON", func(t *testing.T) {

		mockStorage.lastGauge = 123.45
		mockStorage.lastCounter = 10

		gaugeReq := models.Metrics{
			ID:    "test_gauge",
			MType: "gauge",
		}

		body, _ := json.Marshal(gaugeReq)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/value/", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		handler.HandleGetMetricJSON(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var gaugeResp models.Metrics
		err := json.NewDecoder(w.Body).Decode(&gaugeResp)
		assert.NoError(t, err)
		assert.Equal(t, 123.45, *gaugeResp.Value)

		counterReq := models.Metrics{
			ID:    "test_counter",
			MType: "counter",
		}

		body, _ = json.Marshal(counterReq)
		w = httptest.NewRecorder()
		r = httptest.NewRequest("POST", "/value/", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		handler.HandleGetMetricJSON(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var counterResp models.Metrics
		err = json.NewDecoder(w.Body).Decode(&counterResp)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), *counterResp.Delta)
	})
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrInt64(i int64) *int64 {
	return &i
}
