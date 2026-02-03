package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	pprofhttp "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/idudko/go-musthave-metrics/internal/audit"
	"github.com/idudko/go-musthave-metrics/internal/handler"
	"github.com/idudko/go-musthave-metrics/internal/middleware"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/idudko/go-musthave-metrics/internal/service"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func buildInfo(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

func newServer(config Config) (*chi.Mux, repository.Storage, error) {
	var storage repository.Storage
	var pinger handler.DBPinger
	var err error

	if config.DSN != "" {
		dbStorage, err := repository.NewDBStorage(config.DSN)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create database storage: %v", err)
		}
		storage = dbStorage
		pinger = dbStorage
	} else {
		storage, err = repository.NewFileStorage(config.FileStoragePath, config.StoreInterval, config.Restore)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create file storage: %v", err)
		}
	}

	var auditSubject *audit.Subject
	if config.AuditFile != "" || config.AuditURL != "" {
		auditSubject = audit.NewSubject()
	}

	if config.AuditFile != "" {
		fileObserver := audit.NewFileObserver(config.AuditFile)
		auditSubject.Attach(fileObserver)
	}

	if config.AuditURL != "" {
		httpObserver := audit.NewHTTPObserver(config.AuditURL)
		auditSubject.Attach(httpObserver)
	}

	metricsService := service.NewMetricsService(storage)
	h := handler.NewHandler(metricsService, config.Key)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.StripSlashes)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.DecryptionMiddleware(config.CryptoKey))
	r.Use(middleware.HashValidationMiddleware(config.Key))
	r.Use(middleware.GzipRequestMiddleware)
	r.Use(chimiddleware.Compress(5, "application/json", "text/html"))

	if config.AuditFile != "" || config.AuditURL != "" {
		r.Use(middleware.AuditMiddleware(auditSubject))
	}

	r.Post("/update", h.UpdateMetricJSONHandler)
	r.Post("/updates", h.UpdateMetricsBatchHandler)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Post("/value", h.GetMetricValueJSONHandler)
	r.Get("/value/{type}/{name}", h.GetMetricValueHandler)
	r.Get("/", h.ListMetricsHandler)

	pprofRouter := chi.NewRouter()
	pprofRouter.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/debug/pprof/heap", http.StatusTemporaryRedirect)
	}))
	pprofRouter.Get("/heap", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprofhttp.Handler("heap").ServeHTTP(w, r)
	}))
	pprofRouter.Get("/allocs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprofhttp.Handler("allocs").ServeHTTP(w, r)
	}))
	pprofRouter.Get("/goroutine", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprofhttp.Handler("goroutine").ServeHTTP(w, r)
	}))
	pprofRouter.Get("/block", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprofhttp.Handler("block").ServeHTTP(w, r)
	}))
	r.Mount("/debug/pprof", pprofRouter)

	if pinger != nil {
		pingHandler := handler.NewPingHandler(pinger)
		r.Get("/ping", pingHandler.PingHandler)
	}

	return r, storage, nil
}

func main() {
	if err := Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	fmt.Printf("Build version: %s\n", buildInfo(buildVersion))
	fmt.Printf("Build date: %s\n", buildInfo(buildDate))
	fmt.Printf("Build commit: %s\n", buildInfo(buildCommit))

	r, storage, err := newServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if closer, ok := storage.(io.Closer); ok {
		defer closer.Close()
	}

	fmt.Printf("Server is running on %s\n", config.Address)
	if config.Key != "" {
		fmt.Println("Hash validation enabled")
	}

	if config.AuditFile != "" {
		fmt.Printf("Audit file: %s\n", config.AuditFile)
	}

	if config.AuditURL != "" {
		fmt.Printf("Audit URL: %s\n", config.AuditURL)
	}

	if config.CryptoKey != "" {
		fmt.Printf("Crypto key: %s\n", config.CryptoKey)
	}

	if ConfigFile() != "" {
		fmt.Printf("Config file: %s\n", ConfigFile())
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    config.Address,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Setup signal notification for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	// Create shutdown context with timeout
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownRelease()

	// Gracefully shutdown HTTP server
	log.Println("Shutting down HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	// Save metrics before shutdown (for file storage)
	if fileStorage, ok := storage.(interface {
		Save(ctx context.Context) error
	}); ok {
		log.Println("Saving metrics to file...")
		if err := fileStorage.Save(context.Background()); err != nil {
			log.Printf("Error saving metrics: %v", err)
		} else {
			log.Println("Metrics saved successfully")
		}
	}

	// Close storage connection
	if closer, ok := storage.(io.Closer); ok {
		log.Println("Closing storage...")
		if err := closer.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
		} else {
			log.Println("Storage closed successfully")
		}
	}

	log.Println("Server shutdown complete")
}
