package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/repository"
)

var (
	// ErrInvalidMetricType is returned when metric type is neither "gauge" nor "counter".
	ErrInvalidMetricType = errors.New("invalid metric type")
	// ErrInvalidValue is returned when value cannot be converted to expected type.
	ErrInvalidValue = errors.New("invalid value")
	// ErrMetricNotFound is returned when a requested metric does not exist in storage.
	ErrMetricNotFound = errors.New("metric not found")
)

// MetricsService provides business logic for metrics management.
//
// This service acts as an intermediary between HTTP handlers and storage layer,
// handling business logic such as metric validation and type conversion.
type MetricsService struct {
	storage repository.Storage // Storage implementation (mem, file, or database)
}

// NewMetricsService creates a new MetricsService instance.
//
// Parameters:
//   - storage: Storage implementation (mem, file, or database)
//
// Returns:
//   - *MetricsService: Configured service instance
//
// Example:
//
//	storage := repository.NewMemStorage()
//	service := service.NewMetricsService(storage)
func NewMetricsService(storage repository.Storage) *MetricsService {
	return &MetricsService{storage: storage}
}

// UpdateMetric updates a single metric value in storage.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - metricType: Type of metric ("counter" or "gauge")
//   - metricName: Unique metric identifier
//   - metricValue: Value to store (float64 for gauge, int64 for counter)
//
// Returns:
//   - error: ErrInvalidMetricType, ErrInvalidValue, or storage error
//
// Example:
//
//	err := service.UpdateMetric(ctx, "gauge", "cpu_usage", 75.5)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (s *MetricsService) UpdateMetric(ctx context.Context, metricType, metricName string, metricValue any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	switch metricType {
	case model.Gauge:
		if value, ok := metricValue.(float64); ok {
			s.storage.UpdateGauge(ctx, metricName, value)
		} else {
			return ErrInvalidValue
		}
	case model.Counter:
		if value, ok := metricValue.(int64); ok {
			s.storage.UpdateCounter(ctx, metricName, value)
		} else {
			return ErrInvalidValue
		}
	default:
		return ErrInvalidMetricType
	}

	if err := s.storage.Save(ctx); err != nil {
		return err
	}
	return nil
}

// GetMetricValue retrieves the current value of a single metric.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - metricType: Type of metric ("counter" or "gauge")
//   - metricName: Unique metric identifier
//
// Returns:
//   - any: Current metric value (float64 for gauge, int64 for counter)
//   - error: ErrMetricNotFound if metric doesn't exist, ErrInvalidMetricType
//
// Example:
//
//	value, err := service.GetMetricValue(ctx, "gauge", "cpu_usage")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("CPU usage: %.2f%%\n", value.(float64))
func (s *MetricsService) GetMetricValue(ctx context.Context, metricType, metricName string) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	switch metricType {
	case model.Counter:
		counters, err := s.storage.GetCounters(ctx)
		if err != nil {
			return nil, err
		}
		log.Printf("Available counters: %v", counters[metricName])
		value, exists := counters[metricName]
		if !exists {
			return nil, ErrMetricNotFound
		}
		return value, nil
	case model.Gauge:
		gauges, err := s.storage.GetGauges(ctx)
		if err != nil {
			return nil, err
		}
		value, exists := gauges[metricName]
		if !exists {
			return nil, ErrMetricNotFound
		}
		return value, nil
	default:
		return nil, ErrInvalidMetricType
	}
}

// GetGauges retrieves all gauge metrics.
//
// Parameters:
//   - ctx: Context for request cancellation
//
// Returns:
//   - map[string]float64: All gauge metrics with their values
//   - error: Storage error
//
// Example:
//
//	gauges, err := service.GetGauges(ctx)
//	for name, value := range gauges {
//	    fmt.Printf("%s: %.2f\n", name, value)
//	}
func (s *MetricsService) GetGauges(ctx context.Context) (map[string]float64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return s.storage.GetGauges(ctx)
}

// GetCounters retrieves all counter metrics.
//
// Parameters:
//   - ctx: Context for request cancellation
//
// Returns:
//   - map[string]int64: All counter metrics with their values
//   - error: Storage error
//
// Example:
//
//	counters, err := service.GetCounters(ctx)
//	for name, value := range counters {
//	    fmt.Printf("%s: %d\n", name, value)
//	}
func (s *MetricsService) GetCounters(ctx context.Context) (map[string]int64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return s.storage.GetCounters(ctx)
}

// UpdateMetricsBatch updates multiple metrics in a single request.
//
// This is more efficient than multiple individual UpdateMetric calls as it reduces
// HTTP overhead, lock contention, and allows for batch storage operations.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - metrics: Array of metrics to update
//
// Returns:
//   - error: Invalid metric data, missing required fields, or storage error
//
// Example:
//
//	metrics := []model.Metrics{
//		{ID: "cpu", MType: "gauge", Value: ptr.Float64(45.5)},
//		{ID: "requests", MType: "counter", Delta: ptr.Int64(10)},
//	}
//	err := service.UpdateMetricsBatch(ctx, metrics)
//
// Benefits of batch updates:
//   - Reduces HTTP overhead compared to multiple individual requests
//   - Improves performance for sending multiple metrics at once
//   - Minimizes database transaction overhead
func (s *MetricsService) UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error {

	type batchStorage interface {
		UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error
	}

	if batchStorage, ok := s.storage.(batchStorage); ok {
		return batchStorage.UpdateMetricsBatch(ctx, metrics)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, metric := range metrics {
		var err error
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("%s metric value is nil", metric.MType)
			}
			err = s.storage.UpdateGauge(ctx, metric.ID, *metric.Value)
		case model.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("%s metric delta is nil", metric.MType)
			}
			err = s.storage.UpdateCounter(ctx, metric.ID, *metric.Delta)
		default:
			return fmt.Errorf("invalid metric type: %s", metric.MType)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
