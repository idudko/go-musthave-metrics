package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/idudko/go-musthave-metrics/internal/service"
)

// Example_updateMetricViaURL demonstrates updating a metric via URL parameters.
//
// Endpoint: POST /update/{type}/{name}/{value}
//
// Example request:
//
//	POST /update/gauge/cpu_usage/75.5
func Example_updateMetricViaURL() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	req := httptest.NewRequest("POST", "/update/gauge/cpu_usage/75.5", nil)
	w := httptest.NewRecorder()

	// Use chi router to match URL parameters
	router := http.NewServeMux()
	router.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricHandler(w, r)
	})
	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	// Output: Status: 200 OK
}

// Example_updateCounterViaURL demonstrates updating a counter metric via URL parameters.
//
// Endpoint: POST /update/counter/{name}/{value}
//
// Counter metrics are cumulative - each update adds to the existing value.
func Example_updateCounterViaURL() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	// First update
	req1 := httptest.NewRequest("POST", "/update/counter/requests_total/5", nil)
	w1 := httptest.NewRecorder()
	router := http.NewServeMux()
	router.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricHandler(w, r)
	})
	router.ServeHTTP(w1, req1)

	// Second update - counter should now be 10
	req2 := httptest.NewRequest("POST", "/update/counter/requests_total/5", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Get current value
	req3 := httptest.NewRequest("GET", "/value/counter/requests_total", nil)
	w3 := httptest.NewRecorder()
	router.HandleFunc("/value/", func(w http.ResponseWriter, r *http.Request) {
		h.GetMetricValueHandler(w, r)
	})
	router.ServeHTTP(w3, req3)

	body, _ := io.ReadAll(w3.Body)
	fmt.Printf("Counter value: %s\n", body)
	// Output: Counter value: 10
}

// Example_updateMetricViaJSON demonstrates updating a metric via JSON request body.
//
// Endpoint: POST /update/
//
// Request format:
//
//	{
//	  "id": "metric_name",
//	  "type": "gauge",
//	  "value": 75.5
//	}
func Example_updateMetricViaJSON() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	metric := model.Metrics{
		ID:    "cpu_usage",
		MType: model.Gauge,
		Value: float64Ptr(75.5),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := http.NewServeMux()
	router.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricJSONHandler(w, r)
	})
	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var response model.Metrics
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Metric: %s = %v\n", response.ID, *response.Value)
	// Output:
	// Status: 200 OK
	// Metric: cpu_usage = 75.5
}

// Example_updateMetricsBatch demonstrates updating multiple metrics in a single request.
//
// Endpoint: POST /updates/
//
// This is more efficient than multiple individual requests.
func Example_updateMetricsBatch() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	metrics := []model.Metrics{
		{ID: "cpu_usage", MType: model.Gauge, Value: float64Ptr(75.5)},
		{ID: "memory_usage", MType: model.Gauge, Value: float64Ptr(1024.0)},
		{ID: "requests_total", MType: model.Counter, Delta: int64Ptr(10)},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := http.NewServeMux()
	router.HandleFunc("/updates/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricsBatchHandler(w, r)
	})
	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Updated %d metrics\n", len(metrics))
	// Output:
	// Status: 200 OK
	// Updated 3 metrics
}

// Example_getMetricViaURL demonstrates retrieving a metric value via URL parameters.
//
// Endpoint: GET /value/{type}/{name}
func Example_getMetricViaURL() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	// First, set a metric value
	req1 := httptest.NewRequest("POST", "/update/gauge/cpu_usage/75.5", nil)
	w1 := httptest.NewRecorder()
	router := http.NewServeMux()
	router.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricHandler(w, r)
	})
	router.ServeHTTP(w1, req1)

	// Now retrieve it
	req2 := httptest.NewRequest("GET", "/value/gauge/cpu_usage", nil)
	w2 := httptest.NewRecorder()
	router.HandleFunc("/value/", func(w http.ResponseWriter, r *http.Request) {
		h.GetMetricValueHandler(w, r)
	})
	router.ServeHTTP(w2, req2)

	resp := w2.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Value: %s\n", body)
	// Output:
	// Status: 200 OK
	// Value: 75.5
}

// Example_getMetricViaJSON demonstrates retrieving a metric value via JSON request.
//
// Endpoint: POST /value/
//
// Request format:
//
//	{
//	  "id": "metric_name",
//	  "type": "gauge"
//	}
func Example_getMetricViaJSON() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	// First, set a metric value
	req1 := httptest.NewRequest("POST", "/update/gauge/cpu_usage/75.5", nil)
	w1 := httptest.NewRecorder()
	router := http.NewServeMux()
	router.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricHandler(w, r)
	})
	router.ServeHTTP(w1, req1)

	// Now retrieve it via JSON
	query := model.Metrics{
		ID:    "cpu_usage",
		MType: model.Gauge,
	}
	body, _ := json.Marshal(query)

	req2 := httptest.NewRequest("POST", "/value/", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.HandleFunc("/value/", func(w http.ResponseWriter, r *http.Request) {
		h.GetMetricValueJSONHandler(w, r)
	})
	router.ServeHTTP(w2, req2)

	resp := w2.Result()
	defer resp.Body.Close()

	var response model.Metrics
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Value: %v\n", *response.Value)
	// Output:
	// Status: 200 OK
	// Value: 75.5
}

// Example_listAllMetrics demonstrates retrieving all metrics in HTML format.
//
// Endpoint: GET /
//
// Returns an HTML page with all gauge and counter metrics.
func Example_listAllMetrics() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	// Set some metrics
	metrics := []model.Metrics{
		{ID: "cpu_usage", MType: model.Gauge, Value: float64Ptr(75.5)},
		{ID: "memory_usage", MType: model.Gauge, Value: float64Ptr(1024.0)},
		{ID: "requests_total", MType: model.Counter, Delta: int64Ptr(42)},
	}
	body, _ := json.Marshal(metrics)

	req1 := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router := http.NewServeMux()
	router.HandleFunc("/updates/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricsBatchHandler(w, r)
	})
	router.ServeHTTP(w1, req1)

	// List all metrics
	req2 := httptest.NewRequest("GET", "/", nil)
	w2 := httptest.NewRecorder()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		h.ListMetricsHandler(w, r)
	})
	router.ServeHTTP(w2, req2)

	resp := w2.Result()
	defer resp.Body.Close()

	htmlBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Contains HTML: %v\n", bytes.Contains(htmlBody, []byte("<html>")))
	fmt.Printf("Contains cpu_usage: %v\n", bytes.Contains(htmlBody, []byte("cpu_usage")))
	fmt.Printf("Contains requests_total: %v\n", bytes.Contains(htmlBody, []byte("requests_total")))
	// Output:
	// Status: 200 OK
	// Contains HTML: true
	// Contains cpu_usage: true
	// Contains requests_total: true
}

// Example_updateMetricWithHash demonstrates updating a metric with hash-based signing.
//
// When a signing key is configured, the handler validates request hashes
// and returns response hashes in the "HashSHA256" header.
func Example_updateMetricWithHash() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	key := "my-secret-key"
	h := NewHandler(metricsService, key)

	metric := model.Metrics{
		ID:    "cpu_usage",
		MType: model.Gauge,
		Value: float64Ptr(75.5),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := http.NewServeMux()
	router.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricJSONHandler(w, r)
	})
	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("HashSHA256 header present: %v\n", resp.Header.Get("HashSHA256") != "")
	// Output:
	// Status: 200 OK
	// HashSHA256 header present: true
}

// Example_metricNotFound demonstrates handling of non-existent metrics.
//
// Attempting to retrieve a metric that doesn't exist returns 404 Not Found.
func Example_metricNotFound() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	req := httptest.NewRequest("GET", "/value/gauge/nonexistent_metric", nil)
	w := httptest.NewRecorder()

	router := http.NewServeMux()
	router.HandleFunc("/value/", func(w http.ResponseWriter, r *http.Request) {
		h.GetMetricValueHandler(w, r)
	})
	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	// Output: Status: 404 Not Found
}

// Example_invalidMetricType demonstrates handling of invalid metric types.
//
// Using an invalid metric type returns 400 Bad Request.
func Example_invalidMetricType() {
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	h := NewHandler(metricsService, "")

	req := httptest.NewRequest("POST", "/update/unknown_type/metric/10", nil)
	w := httptest.NewRecorder()

	router := http.NewServeMux()
	router.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		h.UpdateMetricHandler(w, r)
	})
	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	// Output: Status: 400 Bad Request
}
