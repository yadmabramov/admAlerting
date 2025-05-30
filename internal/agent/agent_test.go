package agent

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yadmabramov/admAlerting/internal/models"
)

func TestAgent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		gz, err := gzip.NewReader(r.Body)
		assert.NoError(t, err)
		defer gz.Close()

		var metric models.Metrics
		err = json.NewDecoder(gz).Decode(&metric)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(metric)
	}))
	defer ts.Close()

	config := Config{
		ServerURL:      ts.URL,
		PollInterval:   100 * time.Millisecond,
		ReportInterval: 200 * time.Millisecond,
	}

	t.Run("NewAgent initialization", func(t *testing.T) {
		a := NewAgent(config)

		assert.NotNil(t, a.client)
		assert.Equal(t, config.ServerURL, a.serverURL)
		assert.Equal(t, config.PollInterval, a.pollInterval)
		assert.Equal(t, config.ReportInterval, a.reportInterval)
		assert.NotNil(t, a.metrics)
	})

	t.Run("Metric collection", func(t *testing.T) {
		a := NewAgent(config)
		a.collectMetrics()

		assert.NotEmpty(t, a.metrics["Alloc"])
		assert.NotEmpty(t, a.metrics["HeapAlloc"])
		assert.NotEmpty(t, a.metrics["RandomValue"])
		assert.Equal(t, int64(1), a.pollCount)
	})

	t.Run("Metric sending via JSON with gzip", func(t *testing.T) {
		a := NewAgent(config)
		a.collectMetrics()

		err := a.sendMetricJSON("gauge", "TestMetric", "123.45")
		assert.NoError(t, err)

		err = a.sendMetricJSON("counter", "PollCount", "10")
		assert.NoError(t, err)
	})

	t.Run("Send invalid metric via JSON", func(t *testing.T) {
		a := NewAgent(Config{ServerURL: "http://invalid-url"})
		err := a.sendMetricJSON("gauge", "TestMetric", "123.45")
		assert.Error(t, err)
	})

	t.Run("Run agent with intervals using JSON API and gzip", func(t *testing.T) {
		a := NewAgent(config)
		done := make(chan struct{})

		go func() {
			time.Sleep(500 * time.Millisecond)
			close(done)
		}()

		go a.Run()

		select {
		case <-done:
			a.mu.Lock()
			defer a.mu.Unlock()
			assert.True(t, a.pollCount > 0)
			assert.NotEmpty(t, a.metrics)
		case <-time.After(1 * time.Second):
			t.Fatal("Test timeout")
		}
	})
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"integer", 123.0, "123"},
		{"decimal", 123.456, "123.456"},
		{"small value", 0.000123, "0.000123"},
		{"large value", 123456789.0, "123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFloat(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSendMetricJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		gz, err := gzip.NewReader(r.Body)
		assert.NoError(t, err)
		defer gz.Close()

		var metric models.Metrics
		err = json.NewDecoder(gz).Decode(&metric)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metric)
	}))
	defer ts.Close()

	config := Config{
		ServerURL:      ts.URL,
		PollInterval:   100 * time.Millisecond,
		ReportInterval: 200 * time.Millisecond,
	}

	a := NewAgent(config)

	t.Run("Send gauge metric with gzip", func(t *testing.T) {
		err := a.sendMetricJSON("gauge", "TestGauge", "123.45")
		assert.NoError(t, err)
	})

	t.Run("Send counter metric with gzip", func(t *testing.T) {
		err := a.sendMetricJSON("counter", "TestCounter", "10")
		assert.NoError(t, err)
	})

	t.Run("Send invalid metric type", func(t *testing.T) {
		err := a.sendMetricJSON("invalid", "Test", "123")
		assert.Error(t, err)
	})

	t.Run("Send invalid metric value", func(t *testing.T) {
		err := a.sendMetricJSON("gauge", "Test", "not-a-number")
		assert.Error(t, err)
	})
}
