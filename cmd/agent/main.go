package main

import (
	"log"
	"time"

	"github.com/idudko/go-musthave-metrics/internal/agent"
)

const (
	serverAddr     = "http://localhost:8080"
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	collector := agent.NewCollector()

	for {
		collector.Collect()
		time.Sleep(pollInterval)

		err := collector.Report(serverAddr)
		if err != nil {
			log.Printf("error reporting metrics: %v", err)
		}
		time.Sleep(reportInterval - pollInterval)

	}
}
