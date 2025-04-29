package server

import (
	"net/http"

	"github.com/yadmabramov/admAlerting/internal/handlers"
	"github.com/yadmabramov/admAlerting/internal/storage"
)

func NewServer(addr string) *http.Server {
	storage := storage.NewMemoryStorage()
	handler := handlers.NewMetricsHandler(storage)

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.HandleUpdate)
	mux.HandleFunc("/metrics", handler.HandleGetAllMetrics)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
