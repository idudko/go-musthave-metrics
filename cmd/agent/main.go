package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/idudko/go-musthave-metrics/internal/agent"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	UseBatch       bool   `env:"BATCH"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

var config = Config{
	Address:        "localhost:8080",
	PollInterval:   2,
	ReportInterval: 10,
	UseBatch:       true,
	Key:            "",
	RateLimit:      1,
}

func main() {
	if err := cleanenv.ReadEnv(&config); err != nil {
		log.Fatalf("Failed to read config from env: %v", err)
	}

	fset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fset.StringVar(&config.Address, "a", config.Address, "HTTP address to listen on")
	fset.IntVar(&config.PollInterval, "p", config.PollInterval, "Poll interval in seconds")
	fset.IntVar(&config.ReportInterval, "r", config.ReportInterval, "Report interval in seconds")
	fset.BoolVar(&config.UseBatch, "b", config.UseBatch, "Use batch reporting")
	fset.StringVar(&config.Key, "k", config.Key, "Key for signing requests")
	fset.IntVar(&config.RateLimit, "l", config.RateLimit, "Rate limit for concurrent requests")
	fset.Usage = cleanenv.FUsage(fset.Output(), &config, nil, fset.Usage)
	fset.Parse(os.Args[1:])

	if config.RateLimit <= 0 {
		config.RateLimit = 1
	}

	metricsService := agent.NewMetricsService(config.Address, config.Key, config.UseBatch, config.RateLimit)
	metricsService.Start(config.PollInterval, config.ReportInterval)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	metricsService.Stop()
	log.Println("Agent gracefully stopped")
}
