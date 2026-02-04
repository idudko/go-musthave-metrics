package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/idudko/go-musthave-metrics/internal/agent"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// buildInfo returns value or "N/A" if empty
func buildInfo(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

func main() {
	// Initialize configuration from all sources
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 1
	}

	if cfg.ConfigFile != "" {
		log.Printf("Config file: %s", cfg.ConfigFile)
	}

	metricsService := agent.NewMetricsService(cfg.Address, cfg.GrpcAddress, cfg.Key, cfg.UseBatch, cfg.RateLimit, cfg.CryptoKey)
	metricsService.Start(cfg.PollInterval, cfg.ReportInterval)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigChan

	log.Println("Received shutdown signal, gracefully stopping...")

	// Stop collection and send all pending metrics
	metricsService.Shutdown()

	log.Println("Agent gracefully stopped")
}
