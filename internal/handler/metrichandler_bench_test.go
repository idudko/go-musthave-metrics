package handler

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/idudko/go-musthave-metrics/internal/service"

	"github.com/goccy/go-json"
)

// Benchmarks
func BenchmarkHandler_UpdateMetricHandler(b *testing.B) {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService, "")

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/update/gauge/test_metric/123.45", nil)
		handler.UpdateMetricHandler(w, r.WithContext(ctx))
	}
}

func BenchmarkHandler_UpdateMetricJSONHandler(b *testing.B) {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService, "")

	ctx := context.Background()
	metric := model.Metrics{
		ID:    "test_metric",
		MType: "gauge",
		Value: float64Ptr(123.45),
	}

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(metric)
		r := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		handler.UpdateMetricJSONHandler(w, r.WithContext(ctx))
	}
}

func BenchmarkHandler_GetMetricValueHandler(b *testing.B) {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService, "")

	ctx := context.Background()

	// Pre-populate storage
	storage.UpdateGauge(ctx, "test_metric", 123.45)

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/value/gauge/test_metric", nil)
		handler.GetMetricValueHandler(w, r.WithContext(ctx))
	}
}

func BenchmarkHandler_GetMetricValueJSONHandler(b *testing.B) {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService, "")

	ctx := context.Background()

	// Pre-populate storage
	storage.UpdateGauge(ctx, "test_metric", 123.45)

	metric := model.Metrics{
		ID:    "test_metric",
		MType: "gauge",
	}

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(metric)
		r := httptest.NewRequest("POST", "/value/", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		handler.GetMetricValueJSONHandler(w, r.WithContext(ctx))
	}
}

func BenchmarkHandler_UpdateMetricsBatchHandler_Small(b *testing.B) {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService, "")

	ctx := context.Background()

	// Create small batch (10 metrics)
	metrics := make([]model.Metrics, 10)
	for i := range 10 {
		metrics[i] = model.Metrics{
			ID:    "metric_" + strconv.Itoa(i),
			MType: "gauge",
			Value: float64Ptr(float64(i)),
		}
	}

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(metrics)
		r := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		handler.UpdateMetricsBatchHandler(w, r.WithContext(ctx))
	}
}

func BenchmarkHandler_UpdateMetricsBatchHandler_Medium(b *testing.B) {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService, "")

	ctx := context.Background()

	// Create medium batch (100 metrics)
	metrics := make([]model.Metrics, 100)
	for i := range 100 {
		metrics[i] = model.Metrics{
			ID:    "metric_" + strconv.Itoa(i),
			MType: "gauge",
			Value: float64Ptr(float64(i)),
		}
	}

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(metrics)
		r := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		handler.UpdateMetricsBatchHandler(w, r.WithContext(ctx))
	}
}

func BenchmarkHandler_UpdateMetricsBatchHandler_Large(b *testing.B) {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService, "")

	ctx := context.Background()

	// Create large batch (1000 metrics)
	metrics := make([]model.Metrics, 1000)
	for i := range 1000 {
		metrics[i] = model.Metrics{
			ID:    "metric_" + strconv.Itoa(i),
			MType: "gauge",
			Value: float64Ptr(float64(i)),
		}
	}

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(metrics)
		r := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		handler.UpdateMetricsBatchHandler(w, r.WithContext(ctx))
	}
}

func BenchmarkHandler_ListMetricsHandler(b *testing.B) {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService, "")

	ctx := context.Background()

	// Pre-populate storage with metrics
	for i := range 100 {
		storage.UpdateGauge(ctx, "gauge_"+strconv.Itoa(i), float64(i))
		storage.UpdateCounter(ctx, "counter_"+strconv.Itoa(i), int64(i))
	}

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		handler.ListMetricsHandler(w, r.WithContext(ctx))
	}
}
