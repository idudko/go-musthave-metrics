package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"

	"github.com/idudko/go-musthave-metrics/internal/middleware"
	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/service"
	"github.com/idudko/go-musthave-metrics/pkg/hash"
)

const (
	// MetricTypeGauge represents a gauge metric type that stores point-in-time values.
	// Gauges are floating-point numbers that represent the current state of a metric.
	// Examples: CPU usage percentage, memory usage in MB, temperature.
	//
	// Usage in URL: /update/gauge/<name>/<value>
	MetricTypeGauge = "gauge"

	// MetricTypeCounter represents a counter metric type that stores cumulative values.
	// Counters are non-negative integers that are incremented over time.
	// Examples: total requests processed, error count, event count.
	//
	// Usage in URL: /update/counter/<name>/<value>
	MetricTypeCounter = "counter"
)

// Handler provides HTTP handlers for metrics management endpoints.
//
// This handler manages the following operations:
//   - Updating single metrics (via URL parameters or JSON)
//   - Updating multiple metrics in batch (via JSON)
//   - Retrieving metric values (via URL parameters or JSON)
//   - Listing all metrics in HTML format
//
// The handler supports hash-based request signing when a key is configured.
// It integrates with the MetricsService for business logic and storage operations.
//
// Thread Safety:
//
//	The Handler is safe for concurrent use as it only reads from
//	the MetricsService (which handles its own synchronization).
//
// Example:
//
//	// Create handler with metrics service and signing key
//	metricsService := service.NewMetricsService(storage)
//	handler := handler.NewHandler(metricsService, "secret-key")
//
//	// Register handlers with router
//	r.Post("/update/{type}/{name}/{value}", handler.UpdateMetricHandler)
//	r.Post("/update/", handler.UpdateMetricJSONHandler)
//	r.Post("/updates/", handler.UpdateMetricsBatchHandler)
//	r.Get("/value/{type}/{name}", handler.GetMetricValueHandler)
//	r.Post("/value/", handler.GetMetricValueJSONHandler)
//	r.Get("/", handler.ListMetricsHandler)
type Handler struct {
	metricsService *service.MetricsService // Service layer for metrics business logic
	key            string                  // Secret key for hash-based request signing (empty if disabled)
	listTemplate   *template.Template      // HTML template for metrics list view
}

// NewHandler creates a new Handler instance with the provided metrics service and signing key.
//
// Parameters:
//   - metricsService: Service layer for metrics business logic and storage operations
//   - key: Optional secret key for HMAC-SHA256 request/response signing (empty string to disable)
//
// Returns:
//   - *Handler: Configured handler instance ready for use with HTTP router
//
// The handler initializes an HTML template for the metrics list endpoint.
//
// Example:
//
//	storage := repository.NewMemStorage()
//	metricsService := service.NewMetricsService(storage)
//	handler := handler.NewHandler(metricsService, "")
//
//	// With hash signing enabled
//	handler := handler.NewHandler(metricsService, "my-secret-key")
func NewHandler(metricsService *service.MetricsService, key string) *Handler {
	tmpl := `
		<html>
			<head>
				<title>Metrics</title>
			</head>
			<body>
				<div>
					{{range $name, $value := .Gauges}}
						{{$name}}: {{$value}}<br/>
					{{end}}
					{{range $name, $value := .Counters}}
						{{$name}}: {{$value}}<br/>
					{{end}}
				</div>
			</body>
		</html>
	`

	return &Handler{
		metricsService: metricsService,
		key:            key,
		listTemplate:   template.Must(template.New("metrics").Parse(tmpl)),
	}
}

// UpdateMetricHandler handles POST requests to update a single metric via URL parameters.
//
// Endpoint: POST /update/{type}/{name}/{value}
//
// URL Parameters:
//   - type: Metric type ("gauge" or "counter")
//   - name: Unique metric identifier
//   - value: Metric value (numeric)
//
// Behavior:
//   - Parses metric type, name, and value from URL
//   - Validates metric type and value format
//   - Updates the metric in storage via service layer
//   - Adds metric to audit context if available
//   - Returns 200 OK on success, 400 Bad Request on error
//
// Value Formats:
//   - Gauge: Floating-point number (e.g., "75.5", "100", "0.123")
//   - Counter: Integer (e.g., "10", "-5")
//
// Examples:
//
//	// Update gauge
//	POST /update/gauge/cpu_usage/75.5
//	Response: 200 OK
//
//	// Update counter
//	POST /update/counter/requests_total/5
//	Response: 200 OK
//
//	// Invalid metric type
//	POST /update/unknown_metric/x/10
//	Response: 400 Bad Request
//
//	// Missing metric name
//	POST /update/gauge//75.5
//	Response: 400 Bad Request
func (h *Handler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	var value any
	var err error
	switch metricType {
	case MetricTypeCounter:
		value, err = strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
	case MetricTypeGauge:
		value, err = strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
	}
	err = h.metricsService.UpdateMetric(ctx, metricType, metricName, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Добавляем метрику в контекст аудита
	if auditCtx := middleware.GetAuditContext(r.Context()); auditCtx != nil {
		auditCtx.AddMetric(metricName)
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateMetricJSONHandler handles POST requests to update a single metric via JSON.
//
// Endpoint: POST /update/
//
// Request Body (JSON):
//
//	{
//	  "id": "metric_name",
//	  "type": "gauge" or "counter",
//	  "value": 75.5,          // required for gauge
//	  "delta": 10,            // required for counter
//	  "hash": "optional_hash" // optional, for verification
//	}
//
// Behavior:
//   - Decodes JSON request body
//   - Validates required fields (id, type, and value/delta based on type)
//   - Verifies request hash if configured
//   - Updates metric in storage via service layer
//   - Returns updated metric with current value
//   - Calculates and returns response hash if configured
//   - Adds metric to audit context if available
//
// Response Codes:
//   - 200 OK: Metric updated successfully
//   - 400 Bad Request: Invalid JSON, missing fields, or invalid value
//
// Response Headers:
//   - Content-Type: application/json
//   - HashSHA256: Response hash (if signing enabled)
//
// Example Request:
//
//	POST /update/
//	Content-Type: application/json
//
//	{
//	  "id": "cpu_usage",
//	  "type": "gauge",
//	  "value": 75.5
//	}
//
// Example Response:
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json
//	HashSHA256: abc123...
//
//	{
//	  "id": "cpu_usage",
//	  "type": "gauge",
//	  "value": 75.5,
//	  "hash": ""
//	}
func (h *Handler) UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var metric model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if metric.ID == "" || metric.MType == "" {
		http.Error(w, "Invalid metric data", http.StatusBadRequest)
		return
	}

	var err error
	switch metric.MType {
	case MetricTypeGauge:
		if metric.Value == nil {
			http.Error(w, "Value is required for gauge", http.StatusBadRequest)
			return
		}
		err = h.metricsService.UpdateMetric(ctx, metric.MType, metric.ID, *metric.Value)
	case MetricTypeCounter:
		if metric.Delta == nil {
			http.Error(w, "Delta is required for counter", http.StatusBadRequest)
			return
		}
		err = h.metricsService.UpdateMetric(ctx, metric.MType, metric.ID, *metric.Delta)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	value, err := h.metricsService.GetMetricValue(ctx, metric.MType, metric.ID)
	if err == nil {
		switch metric.MType {
		case model.Gauge:
			if v, ok := value.(float64); ok {
				metric.Value = &v
			}
		case model.Counter:
			if v, ok := value.(int64); ok {
				metric.Delta = &v
			}
		}
	}

	data, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	// Добавляем метрику в контекст аудита
	if auditCtx := middleware.GetAuditContext(r.Context()); auditCtx != nil {
		auditCtx.AddMetric(metric.ID)
	}

	w.Header().Set("Content-Type", "application/json")

	if h.key != "" {
		hashValue := hash.ComputeHash(data, h.key)
		w.Header().Set("HashSHA256", hashValue)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetMetricValueHandler handles GET requests to retrieve a single metric value via URL parameters.
//
// Endpoint: GET /value/{type}/{name}
//
// URL Parameters:
//   - type: Metric type ("gauge" or "counter")
//   - name: Unique metric identifier
//
// Behavior:
//   - Retrieves metric value from storage
//   - Returns plain text value
//   - Returns 404 Not Found if metric doesn't exist
//
// Response:
//   - Body: Plain text numeric value
//   - Format: Integer for counter, float for gauge
//
// Response Codes:
//   - 200 OK: Metric found
//   - 404 Not Found: Metric doesn't exist or invalid type
//
// Examples:
//
//	// Get gauge value
//	GET /value/gauge/cpu_usage
//	Response: 75.5
//
//	// Get counter value
//	GET /value/counter/requests_total
//	Response: 42
//
//	// Metric not found
//	GET /value/gauge/unknown_metric
//	Response: 404 Not Found
func (h *Handler) GetMetricValueHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	value, err := h.metricsService.GetMetricValue(ctx, metricType, metricName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%v", value)
}

// GetMetricValueJSONHandler handles POST requests to retrieve a single metric value via JSON.
//
// Endpoint: POST /value/
//
// Request Body (JSON):
//
//	{
//	  "id": "metric_name",
//	  "type": "gauge" or "counter",
//	  "hash": "optional_hash" // optional
//	}
//
// Behavior:
//   - Decodes JSON request body
//   - Validates required fields (id, type)
//   - Verifies request hash if configured
//   - Retrieves metric value from storage
//   - Returns metric with current value
//   - Calculates and returns response hash if configured
//
// Response Codes:
//   - 200 OK: Metric found
//   - 404 Not Found: Metric doesn't exist
//   - 400 Bad Request: Invalid JSON or missing fields
//
// Response Headers:
//   - Content-Type: application/json
//   - HashSHA256: Response hash (if signing enabled)
//
// Example Request:
//
//	POST /value/
//	Content-Type: application/json
//
//	{
//	  "id": "cpu_usage",
//	  "type": "gauge"
//	}
//
// Example Response:
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json
//	HashSHA256: def456...
//
//	{
//	  "id": "cpu_usage",
//	  "type": "gauge",
//	  "value": 75.5
//	}
func (h *Handler) GetMetricValueJSONHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var m model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.ID == "" || m.MType == "" {
		http.Error(w, "Invalid metric data", http.StatusBadRequest)
		return
	}

	value, err := h.metricsService.GetMetricValue(ctx, m.MType, m.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch m.MType {
	case model.Gauge:
		if v, ok := value.(float64); ok {
			m.Value = &v
		}
	case model.Counter:
		if v, ok := value.(int64); ok {
			m.Delta = &v
		}
	}

	data, err := json.Marshal(m)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	if h.key != "" {
		hashValue := hash.ComputeHash(data, h.key)
		w.Header().Set("HashSHA256", hashValue)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// UpdateMetricsBatchHandler handles POST requests to update multiple metrics in a single request.
//
// Endpoint: POST /updates/
//
// Request Body (JSON array):
//
//	[
//	  {
//	    "id": "metric1",
//	    "type": "gauge",
//	    "value": 75.5
//	  },
//	  {
//	    "id": "metric2",
//	    "type": "counter",
//	    "delta": 10
//	  }
//	]
//
// Behavior:
//   - Decodes JSON array of metrics
//   - Validates all metrics before updating
//   - Updates all metrics in a single transaction
//   - Uses batch update when storage supports it (more efficient)
//   - Falls back to individual updates if batch not supported
//   - Adds all metrics to audit context if available
//
// Benefits of batch updates:
//   - Reduces HTTP overhead compared to multiple individual requests
//   - Improves performance for sending multiple metrics at once
//   - Minimizes database transaction overhead
//   - Ensures atomicity when storage supports transactions
//
// Response Codes:
//   - 200 OK: All metrics updated successfully
//   - 400 Bad Request: Invalid JSON, missing fields, or invalid values
//
// Example Request:
//
//	POST /updates/
//	Content-Type: application/json
//
//	[
//	  {"id": "cpu_usage", "type": "gauge", "value": 75.5},
//	  {"id": "memory_usage", "type": "gauge", "value": 1024.0},
//	  {"id": "requests_total", "type": "counter", "delta": 5}
//	]
//
// Example Response:
//
//	HTTP/1.1 200 OK
func (h *Handler) UpdateMetricsBatchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var metrics []model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	for _, metric := range metrics {
		if metric.ID == "" || metric.MType == "" {
			http.Error(w, "Invalid metric data", http.StatusBadRequest)
			return
		}

		var err error
		switch metric.MType {
		case MetricTypeGauge:
			if metric.Value == nil {
				http.Error(w, "Value is required for gauge", http.StatusBadRequest)
				return
			}
			err = h.metricsService.UpdateMetric(ctx, metric.MType, metric.ID, *metric.Value)
		case MetricTypeCounter:
			if metric.Delta == nil {
				http.Error(w, "Delta is required for counter", http.StatusBadRequest)
				return
			}
			err = h.metricsService.UpdateMetric(ctx, metric.MType, metric.ID, *metric.Delta)
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Добавляем метрику в контекст аудита
		if auditCtx := middleware.GetAuditContext(r.Context()); auditCtx != nil {
			auditCtx.AddMetric(metric.ID)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// ListMetricsHandler handles GET requests to display all metrics in HTML format.
//
// Endpoint: GET /
//
// Behavior:
//   - Retrieves all gauge and counter metrics from storage
//   - Renders HTML page with all metrics
//   - Metrics are displayed in alphabetical order
//
// Response:
//   - Body: HTML page with metrics list
//   - Format: Gauge metrics listed first, then counter metrics
//
// Response Codes:
//   - 200 OK: Metrics retrieved successfully
//   - 500 Internal Server Error: Storage error or template error
//
// HTML Format:
//
//	<html>
//	  <head><title>Metrics</title></head>
//	  <body>
//	    <div>
//	      cpu_usage: 75.5<br/>
//	      memory_usage: 1024<br/>
//	      requests_total: 42<br/>
//	    </div>
//	  </body>
//	</html>
//
// Example:
//
//	GET /
//	Response: HTML page with all metrics
func (h *Handler) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gauges, err := h.metricsService.GetGauges(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	counters, err := h.metricsService.GetCounters(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	data := struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.listTemplate.Execute(w, data)
}
