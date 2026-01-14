package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/idudko/go-musthave-metrics/internal/handler"
	"github.com/idudko/go-musthave-metrics/internal/middleware"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/idudko/go-musthave-metrics/internal/service"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DSN             string `env:"DATABASE_DSN"`
}

var config = Config{
	Address:         "localhost:8080",
	StoreInterval:   300,
	FileStoragePath: "metrics.json",
	Restore:         false,
	DSN:             "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
}

func newServer(config Config) (*chi.Mux, error) {
	storage, err := repository.NewFileStorage(config.FileStoragePath, config.StoreInterval, config.Restore)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), config.DSN)
	if err != nil {
		panic(err)
	}

	metricsService := service.NewMetricsService(storage)
	h := handler.NewHandler(metricsService)

	dbmetricsService := service.NewDBMetricsService(pool)
	dbH := handler.NewDBHandler(dbmetricsService)

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
	r.Get("/ping", dbH.PingHandler)
	r.Get("/", h.ListMetricsHandler)

	return r, nil
}

func main() {

	if err := cleanenv.ReadEnv(&config); err != nil {
		log.Fatalf("Failed to read config from env: %v", err)
	}

	fset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fset.StringVar(&config.Address, "a", config.Address, "HTTP address to listen on")
	fset.IntVar(&config.StoreInterval, "i", config.StoreInterval, "Store interval in seconds (0 = synchronous)")
	fset.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "Path to file storage")
	fset.BoolVar(&config.Restore, "r", config.Restore, "Restore metrics from file")
	fset.StringVar(&config.DSN, "d", config.DSN, "PostgreSQL DSN")

	fset.Usage = cleanenv.FUsage(fset.Output(), &config, nil, fset.Usage)
	fset.Parse(os.Args[1:])

	r, err := newServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Server is running on %s\n", config.Address)
	log.Fatal(http.ListenAndServe(config.Address, r))
}
