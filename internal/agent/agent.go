// admAlerting/internal/agent/agent.go
package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Agent struct {
	client         *http.Client
	serverURL      string
	pollInterval   time.Duration
	reportInterval time.Duration
	metrics        map[string]string
	pollCount      int64
	mu             sync.Mutex
}

type Config struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

const (
	Alloc         = "Alloc"
	BuckHashSys   = "BuckHashSys"
	Frees         = "Frees"
	GCCPUFraction = "GCCPUFraction"
	GCSys         = "GCSys"
	HeapAlloc     = "HeapAlloc"
	HeapIdle      = "HeapIdle"
	HeapInuse     = "HeapInuse"
	HeapObjects   = "HeapObjects"
	HeapReleased  = "HeapReleased"
	HeapSys       = "HeapSys"
	LastGC        = "LastGC"
	Lookups       = "Lookups"
	MCacheInuse   = "MCacheInuse"
	MCacheSys     = "MCacheSys"
	MSpanInuse    = "MSpanInuse"
	MSpanSys      = "MSpanSys"
	Mallocs       = "Mallocs"
	NextGC        = "NextGC"
	NumForcedGC   = "NumForcedGC"
	NumGC         = "NumGC"
	OtherSys      = "OtherSys"
	PauseTotalNs  = "PauseTotalNs"
	StackInuse    = "StackInuse"
	StackSys      = "StackSys"
	Sys           = "Sys"
	TotalAlloc    = "TotalAlloc"
	RandomValue   = "RandomValue"
	PollCount     = "PollCount"
)

func NewAgent(config Config) *Agent {
	return &Agent{
		client:         &http.Client{Timeout: 5 * time.Second},
		serverURL:      config.ServerURL,
		pollInterval:   config.PollInterval,
		reportInterval: config.ReportInterval,
		metrics:        make(map[string]string),
	}
}

func (a *Agent) Run() {
	for {
		// Собираем метрики
		a.collectMetrics()

		// Ждем интервал опроса
		time.Sleep(a.pollInterval)

		// Проверяем, не пришло ли время отправки
		if time.Now().UnixNano()%a.reportInterval.Nanoseconds() < a.pollInterval.Nanoseconds() {
			a.sendMetrics()
		}
	}
}

func (a *Agent) collectMetrics() {
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
}

func (a *Agent) sendMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for name, value := range a.metrics {
		if err := a.sendMetric("gauge", name, value); err != nil {
			log.Printf("Failed to send metric %s: %v", name, err)
		}
	}

	if err := a.sendMetric("counter", PollCount, strconv.FormatInt(a.pollCount, 10)); err != nil {
		log.Printf("Failed to send PollCount: %v", err)
	}
}

func (a *Agent) sendMetric(mType, mName, mValue string) error {
	base, err := url.Parse(a.serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	base.Path = path.Join(base.Path, "update", mType, mName, mValue)
	fullURL := base.String()

	resp, err := a.client.Post(fullURL, "text/plain", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
