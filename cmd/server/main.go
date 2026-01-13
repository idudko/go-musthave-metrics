package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/idudko/go-musthave-metrics/internal/handler"
	"github.com/idudko/go-musthave-metrics/internal/middleware"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/idudko/go-musthave-metrics/internal/service"
)

func main() {
	// ADDRESS
	defaultAddress := "localhost:8080"
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		defaultAddress = envAddress
	}

	address := flag.String("a", defaultAddress, "HTTP address to listen on")

	// STORE_INTERVAL
	defaultStoreInterval := 300
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		var err error
		defaultStoreInterval, err = strconv.Atoi(envStoreInterval)
		if err != nil {
			log.Printf("Invalid STORE_INTERVAL env: %v, using default %v", err, defaultStoreInterval)
		}
	}
	storeInterval := flag.Int("i", defaultStoreInterval, "Store interval in seconds (0 = synchronous)")

	// --- FILE_STORAGE_PATH ---
	defaultFile := "metrics.json"
	if envFile := os.Getenv("FILE_STORAGE_PATH"); envFile != "" {
		defaultFile = envFile
	}
	file := flag.String("f", defaultFile, "Path to storage file")

	// RESTORE
	defaultRestore := false
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		var err error
		defaultRestore, err = strconv.ParseBool(envRestore)
		if err != nil {
			log.Printf("Failed to parse RESTORE environment variable: %v", err)
		}
	}
	restore := flag.Bool("r", defaultRestore, "Restore metrics from file")
	flag.Parse()

	storage, err := repository.NewFileStorage(*file, *storeInterval, *restore)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	metricsService := service.NewMetricsService(storage)
	h := handler.NewHandler(metricsService)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.StripSlashes)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.GzipRequestMiddleware)
	r.Use(chimiddleware.Compress(5, "application/json", "text/html"))
	r.Post("/update", h.UpdateMetricJSONHandler)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Post("/value", h.GetMetricValueJSONHandler)
	r.Get("/value/{type}/{name}", h.GetMetricValueHandler)
	r.Get("/", h.ListMetricsHandler)

	fmt.Printf("Server is running on %s\n", *address)
	log.Fatal(http.ListenAndServe(*address, r))
}
