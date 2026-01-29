package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/idudko/go-musthave-metrics/internal/service"
)

func TestUpdateMetricsHandler(t *testing.T) {
	storage := repository.NewMemStorage()
	service := service.NewMetricsService(storage)
	handler := NewHandler(service, "")

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.UpdateMetricHandler)
	r.Post("/update", handler.UpdateMetricJSONHandler)
	tests := []struct {
		name           string
		url            string
		body           any
		expectedStatus int
	}{
		{
			name:           "valid gauge metric",
			url:            "/update",
			body:           model.Metrics{ID: "Alloc", MType: "gauge", Value: float64Ptr(123.45)},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid counter metric",
			url:            "/update",
			body:           model.Metrics{ID: "PollCount", MType: "counter", Delta: int64Ptr(1)},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid metric type",
			url:            "/update",
			body:           model.Metrics{ID: "Metric", MType: "invalid", Value: float64Ptr(123)},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing metric name",
			url:            "/update",
			body:           model.Metrics{MType: "gauge", Value: float64Ptr(123.45)},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid gauge value",
			url:            "/update",
			body:           model.Metrics{ID: "Alloc", MType: "gauge"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid counter value",
			url:            "/update",
			body:           model.Metrics{ID: "PollCount", MType: "counter"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, tt.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
