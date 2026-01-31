package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/idudko/go-musthave-metrics/internal/service"
)

func TestUpdateMetricsHandler(t *testing.T) {
	storage := repository.NewMemStorage()
	service := service.NewMetricsService(storage)
	handler := NewHandler(service)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.UpdateMetricHandler)
	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "valid gauge metric",
			url:            "/update/gauge/Alloc/123.45",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid counter metric",
			url:            "/update/counter/PollCount/1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid metric type",
			url:            "/update/invalid/Metric/123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing metric name",
			url:            "/update/gauge//123.45",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid gauge value",
			url:            "/update/gauge/Alloc/abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid counter value",
			url:            "/update/counter/PollCount/abc",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
