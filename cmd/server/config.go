package main

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"

	configpkg "github.com/idudko/go-musthave-metrics/internal/config"
)

// JSONConfig represents configuration from JSON file
type JSONConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	AuditFile     string `json:"audit_file"`
	AuditURL      string `json:"audit_url"`
}

// Config represents the full configuration
type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"STORE_FILE"`
	Restore         bool   `env:"RESTORE"`
	DSN             string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	CryptoKey       string `env:"CRYPTO_KEY"`
	configFile      string
}

var config = Config{
	Address:         "localhost:8080",
	StoreInterval:   300,
	FileStoragePath: "",
	Restore:         false,
	DSN:             "",
	Key:             "",
	AuditFile:       "",
	AuditURL:        "",
	CryptoKey:       "",
}

// Init registers flags, parses them, and initializes config from all sources.
// Priority order (lowest to highest):
// 1. Default values
// 2. Config file (if specified via -c/-config or CONFIG env var)
// 3. Environment variables
// 4. Command line flags (highest priority)
// Must be called before using any config values.
func Init() error {
	// Register flags with default values
	flag.StringVar(&config.Address, "a", "localhost:8080", "HTTP address to listen on")
	flag.IntVar(&config.StoreInterval, "i", 300, "Store interval in seconds (0 = synchronous)")
	flag.StringVar(&config.FileStoragePath, "f", "", "Path to file storage")
	flag.BoolVar(&config.Restore, "r", false, "Restore metrics from file")
	flag.StringVar(&config.DSN, "d", "", "PostgreSQL DSN")
	flag.StringVar(&config.Key, "k", "", "Key for signing requests")
	flag.StringVar(&config.AuditFile, "audit-file", "", "Path to audit log file")
	flag.StringVar(&config.AuditURL, "audit-url", "", "URL for audit server")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "Path to private key file for decryption")
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

// ConfigFile returns the path to config file if specified
func ConfigFile() string {
	return config.configFile
}

// applyConfig applies config from JSON file with lower priority than env/flags
// Only applies values if the current value is still the default
func applyConfig(cfg *JSONConfig) {
	// Only apply if current value is default
	if cfg.Address != "" && config.Address == "localhost:8080" {
		config.Address = cfg.Address
	}

	if cfg.StoreInterval != "" && config.StoreInterval == 300 {
		if duration, err := configpkg.ParseDuration(cfg.StoreInterval); err == nil {
			config.StoreInterval = duration
		}
	}

	if cfg.StoreFile != "" && config.FileStoragePath == "" {
		config.FileStoragePath = cfg.StoreFile
	}

	if cfg.DatabaseDSN != "" && config.DSN == "" {
		config.DSN = cfg.DatabaseDSN
	}

	if cfg.AuditFile != "" && config.AuditFile == "" {
		config.AuditFile = cfg.AuditFile
	}

	if cfg.AuditURL != "" && config.AuditURL == "" {
		config.AuditURL = cfg.AuditURL
	}

	if cfg.CryptoKey != "" && config.CryptoKey == "" {
		config.CryptoKey = cfg.CryptoKey
	}

	// For boolean flag, only apply if it's true in config and false in current config
	if cfg.Restore && !config.Restore {
		config.Restore = cfg.Restore
	}
}
