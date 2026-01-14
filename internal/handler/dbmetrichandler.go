package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/idudko/go-musthave-metrics/internal/service"
)

type DBHandler struct {
	dbMetricsService *service.DBMetricsService
	PingHandler      http.HandlerFunc
}

func NewDBHandler(dbMetricsService *service.DBMetricsService) *DBHandler {
	return &DBHandler{dbMetricsService: dbMetricsService, PingHandler: PingHandler(dbMetricsService)}
}

func PingHandler(svc *service.DBMetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		if err := svc.Ping(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
