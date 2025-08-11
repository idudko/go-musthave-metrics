package handler

import (
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
