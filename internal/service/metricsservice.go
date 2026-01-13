package service

import (
	"errors"

	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/repository"
)

type MetricsService struct {
	storage repository.Storage
}

func NewMetricsService(storage repository.Storage) *MetricsService {
	return &MetricsService{storage: storage}
}

func (s *MetricsService) UpdateMetric(metricType, metricName string, metricValue any) error {
	switch metricType {
	case model.Gauge:
		if value, ok := metricValue.(float64); ok {
			s.storage.UpdateGauge(metricName, value)
		} else {
			return errors.New("invalid gauge value")
		}
	case model.Counter:
		if value, ok := metricValue.(int64); ok {
			s.storage.UpdateCounter(metricName, value)
		} else {
			return errors.New("invalid counter value")
		}
	default:
		return errors.New("invalid metric type")
	}

	_ = s.storage.Save()
	return nil
}

func (s *MetricsService) GetMetricValue(metricType, metricName string) (any, error) {
	switch metricType {
	case model.Counter:
		counters := s.storage.GetCounters()
		value, exists := counters[metricName]
		if !exists {
			return nil, errors.New("metric not found")
		}
		return value, nil
	case model.Gauge:
		gauges := s.storage.GetGauges()
		value, exists := gauges[metricName]
		if !exists {
			return nil, errors.New("metric not found")
		}
		return value, nil
	default:
		return nil, errors.New("invalid metric type")
	}
}

func (s *MetricsService) GetGauges() map[string]float64 {
	return s.storage.GetGauges()
}

func (s *MetricsService) GetCounters() map[string]int64 {
	return s.storage.GetCounters()
}
