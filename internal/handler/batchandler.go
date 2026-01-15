package handler

import (
	"context"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/idudko/go-musthave-metrics/internal/model"
)

type BatchHandler interface {
	UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error
}

func (h *Handler) UpdateMetricsBatchHandler(w http.ResponseWriter, r *http.Request) {
	var metrics []model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	for i, metric := range metrics {
		if metric.ID == "" || metric.MType == "" {
			http.Error(w, "Invalid metric", http.StatusBadRequest)
			return
		}
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				http.Error(w, "Value is required", http.StatusBadRequest)
				return
			}
		case model.Counter:
			if metric.Delta == nil {
				http.Error(w, "Delta is required", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		metrics[i] = metric
	}

	if err := h.metricsService.UpdateMetricsBatch(r.Context(), metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
