package service

import (
	"errors"
	"strconv"

	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/repository"
)

type MetricsService struct {
	storage repository.Storage
}

func NewMetricsService(storage repository.Storage) *MetricsService {
	return &MetricsService{storage: storage}
}

func (s *MetricsService) UpdateMetric(metricType, metricName, metricValue string) error {
	switch metricType {
	case model.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return errors.New("invalid counter value")
		}
		s.storage.UpdateCounter(metricName, value)
	case model.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return errors.New("invalid gauge value")
		}
		s.storage.UpdateGauge(metricName, value)
	default:
		return errors.New("invalid metric type")
	}

	return nil
}
