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

// GetConfigFilePath returns the path to the configuration file from the flag or CONFIG environment variable.
//
// This function checks if a config file path was provided via the configFlag parameter.
// If not, it checks the CONFIG environment variable.
//
// Parameters:
//   - configFlag: The value from the config flag (-c/-config)
//
// Returns:
//   - string: The path to the config file, or empty string if not specified
//
// Example:
//
//	configFile := config.GetConfigFilePath("")
//	if configFile != "" {
//	    // load from file
//	}
func GetConfigFilePath(configFlag string) string {
	if configFlag != "" {
		return configFlag
	}
	return os.Getenv("CONFIG")
}

// ApplyStringIfDefault applies string value from JSON config only if current value equals default.
//
// Parameters:
//   - current: Pointer to current config value
//   - defaultValue: Default value to compare against
//   - jsonValue: Value from JSON config
//
// Example:
//
//	ApplyStringIfDefault(&cfg.Address, "localhost:8080", jsonCfg.Address)
func ApplyStringIfDefault(current *string, defaultValue, jsonValue string) {
	if jsonValue != "" && *current == defaultValue {
		*current = jsonValue
	}
}

// ApplyDurationIfDefault parses and applies duration from JSON config only if current value equals default.
//
// Parameters:
//   - current: Pointer to current config value
//   - defaultValue: Default value to compare against
//   - jsonValue: Duration string from JSON config
//
// Example:
//
//	ApplyDurationIfDefault(&cfg.PollInterval, 2, jsonCfg.PollInterval)
func ApplyDurationIfDefault(current *int, defaultValue int, jsonValue string) {
	if jsonValue != "" && *current == defaultValue {
		if duration, err := ParseDuration(jsonValue); err == nil {
			*current = duration
		}
	}
}

// ApplyBoolIfDefault applies boolean value from JSON config only if current value is false and JSON value is true.
//
// Parameters:
//   - current: Pointer to current config value
//   - jsonValue: Boolean value from JSON config
//
// Example:
//
//	ApplyBoolIfDefault(&cfg.Restore, jsonCfg.Restore)
func ApplyBoolIfDefault(current *bool, jsonValue bool) {
	if jsonValue && !*current {
		*current = jsonValue
	}
}
