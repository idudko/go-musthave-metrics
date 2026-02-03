package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// loadConfigFile loads configuration from JSON file.
//
// This function reads a JSON file and unmarshals it into the provided config struct.
//
// Parameters:
//   - path: Path to the JSON configuration file
//   - cfg: Pointer to config struct to unmarshal into
//
// Returns:
//   - error: An error if file cannot be read or JSON is invalid
//
// Example:
//
//	var cfg JSONConfig
//	if err := config.LoadConfigFile("config.json", &cfg); err != nil {
//	    log.Fatal(err)
//	}
func LoadConfigFile(path string, cfg interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// parseDuration parses duration string with optional 's' suffix.
//
// The function accepts strings like "10", "10s" and returns the integer value.
//
// Parameters:
//   - s: Duration string to parse
//
// Returns:
//   - int: Duration in seconds
//   - error: An error if parsing fails
//
// Example:
//
//	duration, err := config.ParseDuration("10s")
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseDuration(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty duration")
	}

	// Remove 's' suffix if present
	s = strings.TrimSuffix(s, "s")

	var duration int
	if _, err := fmt.Sscanf(s, "%d", &duration); err != nil {
		return 0, fmt.Errorf("invalid duration format: %w", err)
	}

	if duration <= 0 {
		return 0, fmt.Errorf("duration must be positive, got %d", duration)
	}

	return duration, nil
}
