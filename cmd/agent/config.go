package main

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"

	configpkg "github.com/idudko/go-musthave-metrics/internal/config"
)

// JSONConfig represents configuration from JSON file
type JSONConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
	GrpcAddress    string `json:"grpc_address"`
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
	GrpcAddress    string `env:"GRPC_ADDRESS"`

	// ConfigFile is the path to the configuration file if specified
	ConfigFile string
}

// NewConfig initializes and returns configuration from all sources.
// Priority order (lowest to highest):
// 1. Default values
// 2. JSON config file (if provided via -c or -config or CONFIG env var)
// 3. Environment variables
// 4. Command line flags (highest priority)
//
// Returns a pointer to the initialized Config structure.
func NewConfig() (*Config, error) {
	cfg := &Config{
		Address:        "localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
		UseBatch:       true,
		Key:            "",
		RateLimit:      1,
		CryptoKey:      "",
		GrpcAddress:    "",
	}

	// Register flags with default values
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP address to listen on")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.BoolVar(&cfg.UseBatch, "b", true, "Use batch reporting")
	flag.StringVar(&cfg.Key, "k", "", "Key for signing requests")
	flag.IntVar(&cfg.RateLimit, "l", 1, "Rate limit for concurrent requests")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Path to public key file for encryption")
	flag.StringVar(&cfg.GrpcAddress, "g", "", "gRPC server address")

	var configFileFlag string
	flag.StringVar(&configFileFlag, "c", "", "Path to config file")
	flag.StringVar(&configFileFlag, "config", "", "Path to config file")
	flag.Parse()

	// Get config file path from flag or environment variable
	cfg.ConfigFile = configpkg.GetConfigFilePath(configFileFlag)

	// Load JSON config if file is specified (lower priority than env/flags)
	if cfg.ConfigFile != "" {
		var jsonCfg JSONConfig
		if err := configpkg.LoadConfigFile(cfg.ConfigFile, &jsonCfg); err != nil {
			return nil, err
		}
		cfg.applyJSONConfig(&jsonCfg)
	}

	// Apply environment variables (higher priority than config file, lower than flags)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyJSONConfig applies config from JSON file with lower priority than env/flags
// Only applies values if the current value is still the default
func (c *Config) applyJSONConfig(cfg *JSONConfig) {
	// Apply JSON config values only if current values are still default
	configpkg.ApplyStringIfDefault(&c.Address, "localhost:8080", cfg.Address)
	configpkg.ApplyDurationIfDefault(&c.PollInterval, 2, cfg.PollInterval)
	configpkg.ApplyDurationIfDefault(&c.ReportInterval, 10, cfg.ReportInterval)
	configpkg.ApplyStringIfDefault(&c.CryptoKey, "", cfg.CryptoKey)
	configpkg.ApplyStringIfDefault(&c.GrpcAddress, "", cfg.GrpcAddress)
}
