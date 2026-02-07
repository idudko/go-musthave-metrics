package main

import (
	"flag"

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
	TrustedSubnet string `json:"trusted_subnet"`
	GrpcAddress   string `json:"grpc_address"`
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
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
	GrpcAddress     string `env:"GRPC_ADDRESS"`

	// ConfigFile is the path to the configuration file if specified
	ConfigFile string
}

// NewConfig initializes and returns configuration from all sources.
// Priority order (lowest to highest):
// 1. Default values
// 2. Config file (if specified via -c/-config or CONFIG env var)
// 3. Environment variables
// 4. Command line flags (highest priority)
//
// Returns a pointer to the initialized Config structure.
func NewConfig() (*Config, error) {
	cfg := &Config{
		Address:         "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "",
		Restore:         false,
		DSN:             "",
		Key:             "",
		AuditFile:       "",
		AuditURL:        "",
		CryptoKey:       "",
		TrustedSubnet:   "",
		GrpcAddress:     "",
	}

	// Register flags with default values
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP address to listen on")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "Store interval in seconds (0 = synchronous)")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "Path to file storage")
	flag.BoolVar(&cfg.Restore, "r", false, "Restore metrics from file")
	flag.StringVar(&cfg.DSN, "d", "", "PostgreSQL DSN")
	flag.StringVar(&cfg.Key, "k", "", "Key for signing requests")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "Path to audit log file")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "URL for audit server")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Path to private key file for decryption")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "Trusted subnet in CIDR notation (e.g., 192.168.1.0/24)")
	flag.StringVar(&cfg.GrpcAddress, "g", "", "gRPC address to listen on")

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
	configpkg.ApplyDurationIfDefault(&c.StoreInterval, 300, cfg.StoreInterval)
	configpkg.ApplyStringIfDefault(&c.FileStoragePath, "", cfg.StoreFile)
	configpkg.ApplyStringIfDefault(&c.DSN, "", cfg.DatabaseDSN)
	configpkg.ApplyStringIfDefault(&c.AuditFile, "", cfg.AuditFile)
	configpkg.ApplyStringIfDefault(&c.AuditURL, "", cfg.AuditURL)
	configpkg.ApplyStringIfDefault(&c.CryptoKey, "", cfg.CryptoKey)
	configpkg.ApplyStringIfDefault(&c.TrustedSubnet, "", cfg.TrustedSubnet)
	configpkg.ApplyStringIfDefault(&c.GrpcAddress, "", cfg.GrpcAddress)
	configpkg.ApplyBoolIfDefault(&c.Restore, cfg.Restore)
}
