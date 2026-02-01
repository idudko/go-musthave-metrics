package model

const (
	// Counter is a metric type that represents a cumulative value.
	// Counters are non-negative integers that are incremented over time.
	// Examples: total requests, error count, event count.
	//
	// Usage:
	//   metric.MType = model.Counter
	Counter = "counter"

	// Gaugeisa metric type that represents a point-in-time value.
	// Gauges are floating-point numbers that represent the current state.
	// Examples: CPU usage, memory usage, temperature.
	//
	// Usage:
	//   metric.MType = model.Gauge
	Gauge = "gauge"
)

// Metrics represents a metric value with optional delta and value fields.
// Only one of Delta (for counters) or Value (for gauges) should be set.
//
// The struct uses pointers for Delta and Value to distinguish between:
//   - Not provided (nil): The field is not set
//   - Zero value (0 or 0.0): The field is explicitly set to zero
//
// This allows the JSON encoder to omit these fields when not provided.
//
// Example (Counter):
//
//	metrics.Metrics{
//		ID:    "requests_total",
//		MType: model.Counter,
//		Delta: ptr.Int64(5),
//	}
//
// Example (Gauge):
//
//	metrics.Metrics{
//		ID:    "cpu_usage",
//		MType: model.Gauge,
//		Value: ptr.Float64(45.5),
//	}
type Metrics struct {
	// ID is the unique identifier for the metric.
	// This should be a descriptive name that identifies what the metric measures.
	//
	// Examples:
	//   - "cpu_usage": Current CPU usage percentage
	//   - "memory_usage": Current memory usage in MB
	//   - "requests_total": Total number of requests processed
	//   - "errors_total": Total number of errors encountered
	ID string `json:"id"`

	// MType specifies the type of the metric.
	// It must be either "counter" (model.Counter) or "gauge" (model.Gauge).
	//
	// Valid values:
	//   - "counter": For cumulative metrics (use Delta field)
	//   - "gauge": For point-in-time metrics (use Value field)
	MType string `json:"type"`

	// Delta is the increment value for counter metrics.
	// This field is only used for metrics of type "counter".
	// A nil value means the field is not set.
	// A value of 0 is valid and will be added to the counter.
	//
	// Examples:
	//   - Delta: 5  - Add 5 to the counter
	//   - Delta: 0  - No change (explicit)
	//   - Delta: nil - Field not provided (use Value instead)
	Delta *int64 `json:"delta,omitempty"`

	// Value is the current value for gauge metrics.
	// This field is only used for metrics of type "gauge".
	// A nil value means the field is not set.
	// The value overwrites any previous value for the metric.
	//
	// Examples:
	//   - Value: 45.5  - Set gauge to 45.5
	//   - Value: 0.0   - Set gauge to zero
	//   - Value: nil    - Field not provided (use Delta instead)
	Value *float64 `json:"value,omitempty"`

	// Hash is the SHA256 hash of the request or response body.
	// This field is used for data integrity verification when a secret key is configured.
	// It is populated by the handler and should be validated by the client.
	//
	// Format: Hexadecimal string (64 characters for SHA256)
	//
	// Example:
	//   - Request: Client sends hash in "HashSHA256" header
	//   - Response: Server includes hash in "HashSHA256" header of response
	//   - Value: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	//
	// This field is optional and may be empty if hash signing is disabled.
	Hash string `json:"hash,omitempty"`
}
