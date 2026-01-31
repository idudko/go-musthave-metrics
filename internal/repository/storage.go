package repository

// Storage defines the interface for metric storage.
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}
