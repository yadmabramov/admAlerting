package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yadmabramov/admAlerting/internal/handlers"
	"github.com/yadmabramov/admAlerting/internal/server/log_middleware"
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

	r.Use(log_middleware.LoggerMiddleware(logger)) // используем напрямую

	r.Get("/", handler.HandleIndex)
	r.Post("/update/{type}/{name}/{value}", handler.HandleUpdate)
	r.Get("/value/{type}/{name}", handler.HandleGetMetric)
	r.Get("/metrics", handler.HandleGetAllMetricsJSON)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
