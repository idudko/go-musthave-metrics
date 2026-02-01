package repository

import "context"

// Storage defines the interface for metric storage operations.
//
// This interface provides methods for storing and retrieving metrics.
// Implementations can use in-memory storage, file-based storage, or database storage.
//
// Thread Safety:
//
//	Implementations must be safe for concurrent use from multiple goroutines.
//
// Example implementations:
//   - repository.NewMemStorage(): In-memory storage (default)
//   - repository.NewFileStorage(path): File-based persistence
//   - repository.NewDBStorage(dsn): PostgreSQL database storage
//
// Example usage:
//
//	storage := repository.NewMemStorage()
//	err := storage.UpdateGauge(ctx, "cpu_usage", 75.5)
//	if err != nil {
//	    log.Fatal(err)
//	}
type Storage interface {
	// UpdateGauge stores or updates a gauge metric value.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout handling
	//   - name: Unique identifier for the metric
	//   - value: Floating-point value to store (overwrites previous value)
	//
	// Returns:
	//   - error: Storage error if operation fails
	//
	// Example:
	//
	//	err := storage.UpdateGauge(ctx, "cpu_usage", 75.5)
	//	if err != nil {
	//	    return err
	//	}
	UpdateGauge(ctx context.Context, name string, value float64) error

	// UpdateCounter increments a counter metric value.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout handling
	//   - name: Unique identifier for the metric
	//   - value: Integer value to add to the counter (can be negative)
	//
	// Returns:
	//   - error: Storage error if operation fails
	//
	// Example:
	//
	//	err := storage.UpdateCounter(ctx, "requests_total", 5)
	//	if err != nil {
	//	    return err
	//	}
	UpdateCounter(ctx context.Context, name string, value int64) error

	// GetGauges retrieves all gauge metrics from storage.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout handling
	//
	// Returns:
	//   - map[string]float64: All gauge metrics with their values (empty if none exist)
	//   - error: Storage error if operation fails
	//
	// Example:
	//
	//	gauges, err := storage.GetGauges(ctx)
	//	if err != nil {
	//	    return err
	//	}
	//	for name, value := range gauges {
	//	    fmt.Printf("%s: %.2f\n", name, value)
	//	}
	GetGauges(ctx context.Context) (map[string]float64, error)

	// GetCounters retrieves all counter metrics from storage.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout handling
	//
	// Returns:
	//   - map[string]int64: All counter metrics with their values (empty if none exist)
	//   - error: Storage error if operation fails
	//
	// Example:
	//
	//	counters, err := storage.GetCounters(ctx)
	//	if err != nil {
	//	    return err
	//	}
	//	for name, value := range counters {
	//	    fmt.Printf("%s: %d\n", name, value)
	//	}
	GetCounters(ctx context.Context) (map[string]int64, error)

	// Save persists current state to permanent storage (if applicable).
	//
	// This method is used for file-based storage to write metrics to disk.
	// In-memory storage implementations may return nil without doing anything.
	// Database implementations typically handle persistence automatically.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout handling
	//
	// Returns:
	//   - error: Storage error if operation fails
	//
	// Example:
	//
	//	// For file-based storage
	//	err := storage.Save(ctx)
	//	if err != nil {
	//	    log.Printf("Failed to save to file: %v", err)
	//	    return err
	//	}
	Save(ctx context.Context) error
}
