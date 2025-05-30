package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yadmabramov/admAlerting/internal/handlers"
	"github.com/yadmabramov/admAlerting/internal/server/gzipmiddleware"
	"github.com/yadmabramov/admAlerting/internal/server/logmiddleware"
	"github.com/yadmabramov/admAlerting/internal/service"
	"github.com/yadmabramov/admAlerting/internal/storage"
	"go.uber.org/zap"
)

func NewServer(addr string) *http.Server {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	storage := storage.NewMemoryStorage()
	service := service.NewMetricsService(storage)
	handler := handlers.NewMetricsHandler(service)

	r := chi.NewRouter()

	r.Use(logmiddleware.LoggerMiddleware(logger))
	r.Use(gzipmiddleware.GzipMiddleware)

	r.Get("/", handler.HandleIndex)
	r.Post("/update/{type}/{name}/{value}", handler.HandleUpdate)
	r.Get("/value/{type}/{name}", handler.HandleGetMetric)
	r.Get("/metrics", handler.HandleGetAllMetricsJSON)

	r.Post("/update/", handler.HandleUpdateJSON)
	r.Post("/value/", handler.HandleGetMetricJSON)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
