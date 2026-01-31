package repository

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type FileStorage struct {
	*MemStorage
	path     string
	interval time.Duration
	syncSave bool
	mu       sync.RWMutex
}

type storageData struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func NewFileStorage(path string, interval int, restore bool) (*FileStorage, error) {
	fs := &FileStorage{
		MemStorage: NewMemStorage(),
		path:       path,
		interval:   time.Duration(interval) * time.Second,
		syncSave:   interval == 0,
	}

	ctx := context.Background()

	if restore {
		if err := fs.restore(ctx); err != nil {
			log.Printf("Warning: could not restore metrics from file: %v", err)
		}
	}
	if interval > 0 {
		fs.startAutoSave(ctx)
	}
	return fs, nil
}

func (f *FileStorage) restore(ctx context.Context) error {
	file, err := os.Open(f.path)
	if err != nil {
		return err
	}
	defer file.Close()

	var data storageData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	for k, v := range data.Gauges {
		f.MemStorage.UpdateGauge(ctx, k, v)
	}
	for k, v := range data.Counters {
		f.MemStorage.UpdateCounter(ctx, k, v)
	}
	return nil
}

func (f *FileStorage) saveMetrics(ctx context.Context) error {

	f.mu.Lock()
	defer f.mu.Unlock()
	tmpfile, err := os.CreateTemp("", "metrics-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	gauges, err := f.GetGauges(ctx)
	if err != nil {
		tmpfile.Close()
		return err
	}
	counters, err := f.GetCounters(ctx)
	if err != nil {
		tmpfile.Close()
		return err
	}

	data := storageData{
		Gauges:   gauges,
		Counters: counters,
	}

	if err := json.NewEncoder(tmpfile).Encode(data); err != nil {
		tmpfile.Close()
		return err
	}

	if err := os.Rename(tmpfile.Name(), f.path); err != nil {
		return err
	}

	return tmpfile.Close()
}

func (f *FileStorage) Save(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if f.syncSave {
		return f.saveMetrics(ctx)
	}
	return nil
}

func (f *FileStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.MemStorage.UpdateGauge(ctx, name, value)
	if f.syncSave {
		if err := f.saveMetrics(ctx); err != nil {
			log.Printf("error saving metrics: %v", err)
		}
	}
	return nil
}

func (f *FileStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.MemStorage.UpdateCounter(ctx, name, value)
	if f.syncSave {
		if err := f.saveMetrics(ctx); err != nil {
			log.Printf("error saving metrics: %v", err)
		}
	}
	return nil
}

func (f *FileStorage) startAutoSave(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(f.interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := f.saveMetrics(ctx); err != nil {
				log.Printf("error saving metrics: %v", err)
			}
		}
	}()
}
