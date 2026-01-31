package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/idudko/go-musthave-metrics/internal/agent"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	UseBatch       bool   `env:"BATCH"`
}

var config = Config{
	Address:        "localhost:8080",
	PollInterval:   2,
	ReportInterval: 10,
	UseBatch:       true,
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

	fset.Usage = cleanenv.FUsage(fset.Output(), &config, nil, fset.Usage)
	fset.Parse(os.Args[1:])

	collector := agent.NewCollector()
	pollsSinceReport := 0

	for {
		collector.Collect()
		time.Sleep(time.Duration(config.PollInterval) * time.Second)

		pollsSinceReport++
		if pollsSinceReport >= (config.ReportInterval / config.PollInterval) {
			err := collector.Report(config.Address)
			if err != nil {
				log.Printf("error reporting metrics: %v", err)
			}
			pollsSinceReport = 0
		}
	}
}
