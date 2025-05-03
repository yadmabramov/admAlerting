package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yadmabramov/admAlerting/internal/handlers"
	"github.com/yadmabramov/admAlerting/internal/storage"
)

func NewServer(addr string) *http.Server {
	storage := storage.NewMemoryStorage()
	handler := handlers.NewMetricsHandler(storage)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Роуты
	r.Get("/", handler.HandleIndex)
	r.Post("/update/{type}/{name}/{value}", handler.HandleUpdate)
	r.Get("/value/{type}/{name}", handler.HandleGetMetric)
	r.Get("/metrics", handler.HandleGetAllMetricsJSON)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
