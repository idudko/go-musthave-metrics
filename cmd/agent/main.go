package main

import (
	"flag"
	"log"
	"time"

	"github.com/idudko/go-musthave-metrics/internal/agent"
)

func main() {
	serverAddr := flag.String("a", "localhost:8080", "HTTP server address")
	pollInterval := flag.Int("p", 2, "Poll interval in seconds")
	reportInterval := flag.Int("r", 10, "Report interval in seconds")
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
