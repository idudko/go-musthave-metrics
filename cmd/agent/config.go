package main

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"

	configpkg "github.com/idudko/go-musthave-metrics/internal/config"
)

// JSONConfig represents configuration from JSON file
type JSONConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

// Config represents the full configuration with all sources
type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	UseBatch       bool   `env:"BATCH"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`

	// Internal field for config file path
	configFile string
}

var config = Config{
	Address:        "localhost:8080",
	PollInterval:   2,
	ReportInterval: 10,
	UseBatch:       true,
	Key:            "",
	RateLimit:      1,
	CryptoKey:      "",
}

// applyConfig applies config from JSON file with lower priority than env/flags
// Only applies values if the current value is still the default
func applyConfig(cfg *JSONConfig) {
	// Only apply if current value is default
	if cfg.Address != "" && config.Address == "localhost:8080" {
		config.Address = cfg.Address
	}

	if cfg.PollInterval != "" && config.PollInterval == 2 {
		if duration, err := configpkg.ParseDuration(cfg.PollInterval); err == nil {
			config.PollInterval = duration
		}
	}

	if cfg.ReportInterval != "" && config.ReportInterval == 10 {
		if duration, err := configpkg.ParseDuration(cfg.ReportInterval); err == nil {
			config.ReportInterval = duration
		}
	}

	if cfg.CryptoKey != "" && config.CryptoKey == "" {
		config.CryptoKey = cfg.CryptoKey
	}
}

// Init registers flags and initializes configuration from all sources.
// Priority order (lowest to highest):
// 1. Default values
// 2. JSON config file (if provided via -c or -config or CONFIG env var)
// 3. Environment variables
// 4. Command line flags (highest priority)
// Must be called before using config values.
func Init() error {
	// Register flags with default values
	flag.StringVar(&config.Address, "a", "localhost:8080", "HTTP address to listen on")
	flag.IntVar(&config.PollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&config.ReportInterval, "r", 10, "Report interval in seconds")
	flag.BoolVar(&config.UseBatch, "b", true, "Use batch reporting")
	flag.StringVar(&config.Key, "k", "", "Key for signing requests")
	flag.IntVar(&config.RateLimit, "l", 1, "Rate limit for concurrent requests")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "Path to public key file for encryption")
	flag.StringVar(&config.configFile, "c", "", "Path to config file")
	flag.StringVar(&config.configFile, "config", "", "Path to config file")
	flag.Parse()

	// Get config file path from flag or environment variable
	configFile := config.configFile
	if configFile == "" {
		configFile = os.Getenv("CONFIG")
	}

	// Load JSON config if file is specified (lower priority than env/flags)
	if configFile != "" {
		config.configFile = configFile
		var jsonCfg JSONConfig
		if err := configpkg.LoadConfigFile(configFile, &jsonCfg); err != nil {
			return err
		}
		applyConfig(&jsonCfg)
	}

	// Apply environment variables (higher priority than config file, lower than flags)
	if err := cleanenv.ReadEnv(&config); err != nil {
		return err
	}

	return nil
}
