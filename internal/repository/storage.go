package repository

import "context"

// Storage defines the interface for metric storage.
type Storage interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	GetGauges(ctx context.Context) (map[string]float64, error)
	GetCounters(ctx context.Context) (map[string]int64, error)
	Save(ctx context.Context) error
}
