package service

import (
	"context"
	"errors"

	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/repository"
)

var (
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrInvalidValue      = errors.New("invalid value")
)

type MetricsService struct {
	storage repository.Storage
}

func NewMetricsService(storage repository.Storage) *MetricsService {
	return &MetricsService{storage: storage}
}

func (s *MetricsService) UpdateMetric(ctx context.Context, metricType, metricName string, metricValue any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	switch metricType {
	case model.Gauge:
		if value, ok := metricValue.(float64); ok {
			s.storage.UpdateGauge(ctx, metricName, value)
		} else {
			return ErrInvalidValue
		}
	case model.Counter:
		if value, ok := metricValue.(int64); ok {
			s.storage.UpdateCounter(ctx, metricName, value)
		} else {
			return ErrInvalidValue
		}
	default:
		return ErrInvalidMetricType
	}

	if err := s.storage.Save(ctx); err != nil {
		return err
	}
	return nil
}

func (s *MetricsService) GetMetricValue(ctx context.Context, metricType, metricName string) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	switch metricType {
	case model.Counter:
		counters, err := s.storage.GetCounters(ctx)
		if err != nil {
			return nil, err
		}
		value, exists := counters[metricName]
		if !exists {
			return nil, errors.New("metric not found")
		}
		return value, nil
	case model.Gauge:
		gauges, err := s.storage.GetGauges(ctx)
		if err != nil {
			return nil, err
		}
		value, exists := gauges[metricName]
		if !exists {
			return nil, errors.New("metric not found")
		}
		return value, nil
	default:
		return nil, errors.New("invalid metric type")
	}
}

func (s *MetricsService) GetGauges(ctx context.Context) (map[string]float64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return s.storage.GetGauges(ctx)
}

func (s *MetricsService) GetCounters(ctx context.Context) (map[string]int64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return s.storage.GetCounters(ctx)
}
