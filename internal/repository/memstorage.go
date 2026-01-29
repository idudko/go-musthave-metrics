package repository

import (
	"context"
	"sync"
	"sync/atomic"
)

type MemStorage struct {
	gauges         map[string]float64
	counters       map[string]int64
	cachedGauges   atomic.Value // stores *map[string]float64
	cachedCounters atomic.Value // stores *map[string]int64
	version        atomic.Int64
	mu             sync.RWMutex
}

func NewMemStorage() *MemStorage {
	initialGauges := make(map[string]float64)
	initialCounters := make(map[string]int64)

	s := &MemStorage{
		gauges:   initialGauges,
		counters: initialCounters,
	}

	// Store initial cached maps
	s.cachedGauges.Store(&initialGauges)
	s.cachedCounters.Store(&initialCounters)
	s.version.Store(0)

	return s
}

func (s *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	s.gauges[name] = value
	s.version.Add(1)
	s.mu.Unlock()
	return nil
}

func (s *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	s.counters[name] += value
	s.version.Add(1)
	s.mu.Unlock()
	return nil
}

func (s *MemStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	// Return cached map - no allocation on read
	if cached := s.cachedGauges.Load(); cached != nil {
		return *cached.(*map[string]float64), nil
	}
	return make(map[string]float64), nil
}

func (s *MemStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	// Return cached map - no allocation on read
	if cached := s.cachedCounters.Load(); cached != nil {
		return *cached.(*map[string]int64), nil
	}
	return make(map[string]int64), nil
}

func (s *MemStorage) Save(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
