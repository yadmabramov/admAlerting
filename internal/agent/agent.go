package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/yadmabramov/admAlerting/internal/models"
	"github.com/yadmabramov/admAlerting/internal/utils"
)

const (
	maxRetries      = 3
	initialDelay    = 1 * time.Second
	maxBackoffDelay = 5 * time.Second
)

type Agent struct {
	client         *http.Client
	serverURL      string
	pollInterval   time.Duration
	reportInterval time.Duration
	metrics        map[string]string
	pollCount      int64
	mu             sync.Mutex
	config         Config
	metricsChan    chan map[string]string
	workerPool     chan struct{}
	ctx            context.Context
	cancel         context.CancelFunc
}

type Config struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
	Key            string
	RateLimit      int
}

const (
	Alloc           = "Alloc"
	BuckHashSys     = "BuckHashSys"
	Frees           = "Frees"
	GCCPUFraction   = "GCCPUFraction"
	GCSys           = "GCSys"
	HeapAlloc       = "HeapAlloc"
	HeapIdle        = "HeapIdle"
	HeapInuse       = "HeapInuse"
	HeapObjects     = "HeapObjects"
	HeapReleased    = "HeapReleased"
	HeapSys         = "HeapSys"
	LastGC          = "LastGC"
	Lookups         = "Lookups"
	MCacheInuse     = "MCacheInuse"
	MCacheSys       = "MCacheSys"
	MSpanInuse      = "MSpanInuse"
	MSpanSys        = "MSpanSys"
	Mallocs         = "Mallocs"
	NextGC          = "NextGC"
	NumForcedGC     = "NumForcedGC"
	NumGC           = "NumGC"
	OtherSys        = "OtherSys"
	PauseTotalNs    = "PauseTotalNs"
	StackInuse      = "StackInuse"
	StackSys        = "StackSys"
	Sys             = "Sys"
	TotalAlloc      = "TotalAlloc"
	RandomValue     = "RandomValue"
	PollCount       = "PollCount"
	TotalMemory     = "TotalMemory"
	FreeMemory      = "FreeMemory"
	CPUutilization1 = "CPUutilization1"
)

func NewAgent(config Config) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		client:         &http.Client{Timeout: 5 * time.Second},
		serverURL:      config.ServerURL,
		pollInterval:   config.PollInterval,
		reportInterval: config.ReportInterval,
		metrics:        make(map[string]string),
		config:         config,
		metricsChan:    make(chan map[string]string, 100),
		workerPool:     make(chan struct{}, config.RateLimit),
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (a *Agent) Run() {

	log.Printf("Starting agent with config (priority: ENV > FLAGS > DEFAULTS):\n"+
		"  Server URL:      %s\n"+
		"  Poll Interval:   %v (%.0f seconds)\n"+
		"  Report Interval: %v (%.0f seconds)\n"+
		"  Rate Limit:      %d\n"+
		"  Using key:       %t",
		a.serverURL,
		a.pollInterval, a.pollInterval.Seconds(),
		a.reportInterval, a.reportInterval.Seconds(),
		a.config.RateLimit,
		a.config.Key != "")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		a.pollMetrics()
	}()

	go func() {
		defer wg.Done()
		a.sendMetrics()
	}()

	<-a.ctx.Done()

	a.cancel()

	wg.Wait()

	log.Println("Agent stopped")
}

func (a *Agent) pollMetrics() {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.collectRuntimeMetrics()
			a.collectGopsutilMetrics()
		case <-a.ctx.Done():
			return // Завершаем горутину при отмене контекста
		}
	}
}

func (a *Agent) collectRuntimeMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	a.metrics[Alloc] = formatFloat(float64(stats.Alloc))
	a.metrics[BuckHashSys] = formatFloat(float64(stats.BuckHashSys))
	a.metrics[Frees] = formatFloat(float64(stats.Frees))
	a.metrics[GCCPUFraction] = formatFloat(stats.GCCPUFraction)
	a.metrics[GCSys] = formatFloat(float64(stats.GCSys))
	a.metrics[HeapAlloc] = formatFloat(float64(stats.HeapAlloc))
	a.metrics[HeapIdle] = formatFloat(float64(stats.HeapIdle))
	a.metrics[HeapInuse] = formatFloat(float64(stats.HeapInuse))
	a.metrics[HeapObjects] = formatFloat(float64(stats.HeapObjects))
	a.metrics[HeapReleased] = formatFloat(float64(stats.HeapReleased))
	a.metrics[HeapSys] = formatFloat(float64(stats.HeapSys))
	a.metrics[LastGC] = formatFloat(float64(stats.LastGC))
	a.metrics[Lookups] = formatFloat(float64(stats.Lookups))
	a.metrics[MCacheInuse] = formatFloat(float64(stats.MCacheInuse))
	a.metrics[MCacheSys] = formatFloat(float64(stats.MCacheSys))
	a.metrics[MSpanInuse] = formatFloat(float64(stats.MSpanInuse))
	a.metrics[MSpanSys] = formatFloat(float64(stats.MSpanSys))
	a.metrics[Mallocs] = formatFloat(float64(stats.Mallocs))
	a.metrics[NextGC] = formatFloat(float64(stats.NextGC))
	a.metrics[NumForcedGC] = formatFloat(float64(stats.NumForcedGC))
	a.metrics[NumGC] = formatFloat(float64(stats.NumGC))
	a.metrics[OtherSys] = formatFloat(float64(stats.OtherSys))
	a.metrics[PauseTotalNs] = formatFloat(float64(stats.PauseTotalNs))
	a.metrics[StackInuse] = formatFloat(float64(stats.StackInuse))
	a.metrics[StackSys] = formatFloat(float64(stats.StackSys))
	a.metrics[Sys] = formatFloat(float64(stats.Sys))
	a.metrics[TotalAlloc] = formatFloat(float64(stats.TotalAlloc))

	a.metrics[RandomValue] = formatFloat(rand.Float64() * 100)
	a.pollCount++

	// Отправляем копию метрик в канал
	metricsCopy := make(map[string]string, len(a.metrics))
	for k, v := range a.metrics {
		metricsCopy[k] = v
	}
	a.metricsChan <- metricsCopy
}

func (a *Agent) collectGopsutilMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Собираем метрики памяти
	if v, err := mem.VirtualMemory(); err == nil {
		a.metrics[TotalMemory] = formatFloat(float64(v.Total))
		a.metrics[FreeMemory] = formatFloat(float64(v.Free))
	}

	// Собираем метрики CPU
	if percents, err := cpu.Percent(time.Second, false); err == nil && len(percents) > 0 {
		a.metrics[CPUutilization1] = formatFloat(percents[0])
	}
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (a *Agent) sendBatchMetrics(metrics map[string]string) error {
	return utils.Retry(maxRetries, initialDelay, func() error {
		var metricsList []models.Metrics

		// Добавляем gauge метрики
		for name, value := range metrics {
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue // Пропускаем невалидные метрики
			}
			metricsList = append(metricsList, models.Metrics{
				ID:    name,
				MType: "gauge",
				Value: &val,
			})
		}

		// Добавляем counter метрику
		a.mu.Lock()
		pollCount := a.pollCount
		a.mu.Unlock()

		if pollCount > 0 {
			metricsList = append(metricsList, models.Metrics{
				ID:    PollCount,
				MType: "counter",
				Delta: &pollCount,
			})
		}

		jsonData, err := json.Marshal(metricsList)
		if err != nil {
			return fmt.Errorf("failed to marshal metrics: %w", err)
		}

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(jsonData); err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
		if err := gz.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}

		url := a.serverURL + "/updates/"
		req, err := http.NewRequest("POST", url, &buf)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")

		if a.config.Key != "" {
			hash := utils.CalculateHash(jsonData, a.config.Key)
			req.Header.Set("HashSHA256", hash)
		}

		resp, err := a.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		var response []models.Metrics
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})
}

func (a *Agent) sendMetricsIndividually(metrics map[string]string) {
	for name, value := range metrics {
		if err := a.sendMetricJSON("gauge", name, value); err != nil {
			log.Printf("Failed to send metric %s: %v", name, err)
		}
	}

	a.mu.Lock()
	pollCount := a.pollCount
	a.mu.Unlock()

	if err := a.sendMetricJSON("counter", PollCount, strconv.FormatInt(pollCount, 10)); err != nil {
		log.Printf("Failed to send PollCount: %v", err)
	}
}

func (a *Agent) sendMetricJSON(mType, mName, mValue string) error {
	return utils.Retry(maxRetries, initialDelay, func() error {

		var metric models.Metrics

		switch mType {
		case "gauge":
			val, err := strconv.ParseFloat(mValue, 64)
			if err != nil {
				return fmt.Errorf("invalid gauge value: %w", err)
			}
			metric = models.Metrics{
				ID:    mName,
				MType: mType,
				Value: &val,
			}
		case "counter":
			val, err := strconv.ParseInt(mValue, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid counter value: %w", err)
			}
			metric = models.Metrics{
				ID:    mName,
				MType: mType,
				Delta: &val,
			}
		default:
			return fmt.Errorf("unknown metric type: %s", mType)
		}

		jsonData, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("failed to marshal metric: %w", err)
		}

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(jsonData); err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
		if err := gz.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}

		url := a.serverURL + "/update/"
		req, err := http.NewRequest("POST", url, &buf)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		if a.config.Key != "" {
			hash := utils.CalculateHash(jsonData, a.config.Key)
			req.Header.Set("HashSHA256", hash)
		}

		resp, err := a.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		var response models.Metrics
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})
}

func (a *Agent) sendMetrics() {
	for {
		select {
		case metrics := <-a.metricsChan:

			a.workerPool <- struct{}{}

			go func(m map[string]string) {
				defer func() { <-a.workerPool }()

				if err := a.sendBatchMetrics(m); err != nil {
					log.Printf("Failed to send batch metrics: %v, falling back to single metric mode", err)
					a.sendMetricsIndividually(m)
				}
			}(metrics)

		case <-a.ctx.Done():
			return
		}
	}
}

func (a *Agent) Shutdown() {
	a.cancel()
}
