package repository

import (
	"context"
	"maps"
	"sync"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
	return nil
}

func (s *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value
	return nil
}

func (s *MemStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.gauges), nil
}

func (s *MemStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.counters), nil
}

func (s *MemStorage) Save(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
