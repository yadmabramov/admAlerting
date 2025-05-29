package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgent(t *testing.T) {
	// Создаем тестовый сервер для мокирования HTTP-запросов
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Конфигурация для тестов
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

	t.Run("Metric sending", func(t *testing.T) {
		a := NewAgent(config)
		a.collectMetrics()

		err := a.sendMetric("gauge", "TestMetric", "123.45")
		assert.NoError(t, err)

		err = a.sendMetric("counter", "PollCount", "10")
		assert.NoError(t, err)
	})

	t.Run("Send invalid metric", func(t *testing.T) {
		a := NewAgent(Config{ServerURL: "http://invalid-url"})
		err := a.sendMetric("gauge", "TestMetric", "123.45")
		assert.Error(t, err)
	})

	t.Run("Run agent with intervals", func(t *testing.T) {
		a := NewAgent(config)
		done := make(chan struct{})

		// Запускаем агент на короткое время
		go func() {
			time.Sleep(500 * time.Millisecond)
			close(done)
		}()

		go a.Run()

		select {
		case <-done:
			// Проверяем, что метрики были собраны и отправлены
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
