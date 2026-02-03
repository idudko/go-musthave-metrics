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
	if err := Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	if config.RateLimit <= 0 {
		config.RateLimit = 1
	}

	if config.configFile != "" {
		log.Printf("Config file: %s", config.configFile)
	}

	metricsService := agent.NewMetricsService(config.Address, config.Key, config.UseBatch, config.RateLimit, config.CryptoKey)
	metricsService.Start(config.PollInterval, config.ReportInterval)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigChan

	log.Println("Received shutdown signal, gracefully stopping...")

	// Stop collection and send all pending metrics
	metricsService.Shutdown()

	log.Println("Agent gracefully stopped")
}
