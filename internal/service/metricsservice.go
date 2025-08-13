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

func (s *MetricsService) GetMetricValue(metricType, metricName string) (interface{}, error) {
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
