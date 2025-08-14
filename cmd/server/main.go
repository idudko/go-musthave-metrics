package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/idudko/go-musthave-metrics/internal/handler"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/idudko/go-musthave-metrics/internal/service"
)

func main() {
	address := flag.String("a", "localhost:8080", "HTTP address to listen on")
	flag.Parse()

	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := handler.NewHandler(metricsService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Get("/value/{type}/{name}", h.GetMetricValueHandler)
	r.Get("/", h.ListMetricsHandler)

	fmt.Printf("Server is running on %s\n", *address)
	log.Fatal(http.ListenAndServe(*address, r))
}
