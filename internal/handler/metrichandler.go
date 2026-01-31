package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	err := h.metricsService.UpdateMetric(metricType, metricName, metricValue)
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
