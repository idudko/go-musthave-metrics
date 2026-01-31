package service

import (
	"context"
	"errors"
	"fmt"

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

func (s *MetricsService) UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error {

	type batchStorage interface {
		UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error
	}

	if batchStorage, ok := s.storage.(batchStorage); ok {
		return batchStorage.UpdateMetricsBatch(ctx, metrics)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, metric := range metrics {
		var err error
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("%s metric value is nil", metric.MType)
			}
			err = s.storage.UpdateGauge(ctx, metric.ID, *metric.Value)
		case model.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("%s metric delta is nil", metric.MType)
			}
			err = s.storage.UpdateCounter(ctx, metric.ID, *metric.Delta)
		default:
			return fmt.Errorf("invalid metric type: %s", metric.MType)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
