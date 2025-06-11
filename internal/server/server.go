package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yadmabramov/admAlerting/internal/handlers"
	"github.com/yadmabramov/admAlerting/internal/server/gzipmiddleware"
	"github.com/yadmabramov/admAlerting/internal/server/logmiddleware"
	"github.com/yadmabramov/admAlerting/internal/service"
	"github.com/yadmabramov/admAlerting/internal/storage"
	"go.uber.org/zap"
)

type Config struct {
	Addr          string
	StoreInterval time.Duration
	StoragePath   string
	Restore       bool
	DatabaseDSN   string
}

type Server struct {
	*http.Server
	config  Config
	storage storage.Repository
	logger  *zap.Logger
	stop    chan struct{}
	wg      sync.WaitGroup
	db      *sql.DB
}

func NewServer(config Config) *Server {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	var repo storage.Repository
	var db *sql.DB

	if config.DatabaseDSN != "" {
		postgresStorage, err := storage.NewPostgresStorage(config.DatabaseDSN)
		if err != nil {
			logger.Fatal("Failed to initialize PostgreSQL storage", zap.Error(err))
		}
		repo = postgresStorage
		db = postgresStorage.GetDB()
		logger.Info("Using PostgreSQL storage")
	} else {
		if config.StoragePath != "" {
			memStorage := storage.NewMemoryStorage()
			repo = memStorage
			if config.Restore {
				if err := loadMetricsFromFile(config.StoragePath, repo); err != nil {
					logger.Error("Failed to load metrics from file", zap.Error(err))
				}
			}
			logger.Info("Using file storage", zap.String("path", config.StoragePath))
		} else {
			repo = storage.NewMemoryStorage()
			logger.Info("Using in-memory storage")
		}
	}

	service := service.NewMetricsService(repo)
	handler := handlers.NewMetricsHandler(service)

	r := chi.NewRouter()
	r.Use(logmiddleware.LoggerMiddleware(logger))
	r.Use(gzipmiddleware.GzipMiddleware)

	r.Get("/", handler.HandleIndex)
	r.Post("/update/{type}/{name}/{value}", handler.HandleUpdate)
	r.Get("/value/{type}/{name}", handler.HandleGetMetric)
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler.HandleGetAllMetricsJSON(w, r)
	})
	r.Post("/update/", handler.HandleUpdateJSON)
	r.Post("/value/", handler.HandleGetMetricJSON)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    config.Addr,
		Handler: r,
	}

	server := &Server{
		Server:  srv,
		config:  config,
		storage: repo,
		logger:  logger,
		stop:    make(chan struct{}),
		db:      db,
	}

	if config.StoragePath != "" && config.DatabaseDSN == "" && config.StoreInterval > 0 {
		server.wg.Add(1)
		go server.startSaver()
	}

	return server
}

func (s *Server) startSaver() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.saveMetrics(); err != nil {
				s.logger.Error("Failed to save metrics", zap.Error(err))
			}
		case <-s.stop:
			if err := s.saveMetrics(); err != nil {
				s.logger.Error("Failed to save metrics on shutdown", zap.Error(err))
			}
			return
		}
	}
}

func (s *Server) saveMetrics() error {
	gauges, counters := s.storage.GetAllMetrics()

	data := struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	// Создаем директорию, если она не существует
	if err := os.MkdirAll(filepath.Dir(s.config.StoragePath), 0755); err != nil {
		return err
	}

	file, err := os.Create(s.config.StoragePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func loadMetricsFromFile(path string, storage storage.Repository) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var data struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	for name, value := range data.Gauges {
		if err := storage.UpdateGauge(name, value); err != nil {
			return err
		}
	}

	for name, value := range data.Counters {
		if err := storage.UpdateCounter(name, value); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) ListenAndServe() error {
	errChan := make(chan error)
	go func() {
		errChan <- s.Server.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		return err
	case <-s.stop:
		return nil
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	close(s.stop)
	s.wg.Wait()

	if s.config.StoreInterval > 0 {
		if err := s.saveMetrics(); err != nil {
			s.logger.Error("Failed to save metrics on shutdown", zap.Error(err))
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.logger.Error("Failed to close database connection", zap.Error(err))
		}
	}

	if err := s.Server.Shutdown(ctx); err != nil {
		return err
	}

	s.logger.Sync()
	return nil
}
