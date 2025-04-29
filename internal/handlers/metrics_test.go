package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yadmabramov/admAlerting/internal/storage"
)

type mockStorage struct {
	storage.Repository
	lastGauge   float64
	lastCounter int64
}

func (m *mockStorage) UpdateGauge(name string, value float64) {
	m.lastGauge = value
}

func (m *mockStorage) UpdateCounter(name string, value int64) {
	m.lastCounter = value
}

func TestMetricsHandler(t *testing.T) {
	mock := &mockStorage{}
	handler := NewMetricsHandler(mock)

	t.Run("Update gauge", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/update/gauge/test/123.45", nil)
		w := httptest.NewRecorder()

		handler.HandleUpdate(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 123.45, mock.lastGauge)
	})

	t.Run("Update counter", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/update/counter/test/10", nil)
		w := httptest.NewRecorder()

		handler.HandleUpdate(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, int64(10), mock.lastCounter)
	})

	t.Run("Invalid method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/update/gauge/test/123.45", nil)
		w := httptest.NewRecorder()

		handler.HandleUpdate(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
