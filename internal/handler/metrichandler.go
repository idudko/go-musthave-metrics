package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/service"
)

type Handler struct {
	metricsService *service.MetricsService
}

func NewHandler(metricsService *service.MetricsService) *Handler {
	return &Handler{metricsService: metricsService}
}

func (h *Handler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
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
	case "counter":
		value, err = strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
	case "gauge":
		value, err = strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
	}
	err = h.metricsService.UpdateMetric(metricType, metricName, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
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
	case "gauge":
		if metric.Value == nil {
			http.Error(w, "Value is required for gauge", http.StatusBadRequest)
			return
		}
		err = h.metricsService.UpdateMetric(metric.MType, metric.ID, *metric.Value)
	case "counter":
		if metric.Delta == nil {
			http.Error(w, "Delta is required for counter", http.StatusBadRequest)
			return
		}
		err = h.metricsService.UpdateMetric(metric.MType, metric.ID, *metric.Delta)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricValueHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	value, err := h.metricsService.GetMetricValue(metricType, metricName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%v", value)
}

func (h *Handler) GetMetricValueJSONHandler(w http.ResponseWriter, r *http.Request) {

	var m model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.ID == "" || m.MType == "" {
		http.Error(w, "Invalid metric data", http.StatusBadRequest)
		return
	}

	value, err := h.metricsService.GetMetricValue(m.MType, m.ID)
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	gauges := h.metricsService.GetGauges()
	counters := h.metricsService.GetCounters()

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

	data := struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	t := template.Must(template.New("metrics").Parse(tmpl))
	t.Execute(w, data)
}
