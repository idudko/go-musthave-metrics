package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/idudko/go-musthave-metrics/internal/agent"
)

func main() {
	defaultAddress := "localhost:8080"
	defaultReportInterval := 10
	defaultPollInterval := 2

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		defaultAddress = envAddress
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		if val, err := strconv.Atoi(envReportInterval); err == nil {
			defaultReportInterval = val
		}
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		if val, err := strconv.Atoi(envPollInterval); err == nil {
			defaultPollInterval = val
		}
	}

	serverAddr := flag.String("a", defaultAddress, "HTTP server address")
	pollInterval := flag.Int("p", defaultPollInterval, "Poll interval in seconds")
	reportInterval := flag.Int("r", defaultReportInterval, "Report interval in seconds")
	flag.Parse()
	collector := agent.NewCollector()

	for {
		collector.Collect()
		time.Sleep(time.Duration(*pollInterval) * time.Second)

		err := collector.Report(*serverAddr)
		if err != nil {
			log.Printf("error reporting metrics: %v", err)
		}
		time.Sleep(time.Duration(*reportInterval-*pollInterval) * time.Second)
	}
}
