package repository

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type FileStorage struct {
	*MemStorage
	path     string
	interval time.Duration
	syncSave bool
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

	if restore {
		if err := fs.restore(); err != nil {
			log.Printf("Warning: could not restore metrics from file: %v", err)
		}
	}
	if interval > 0 {
		fs.startAutoSave()
	}
	return fs, nil
}

func (f *FileStorage) restore() error {
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
		f.MemStorage.UpdateGauge(k, v)
	}
	for k, v := range data.Counters {
		f.MemStorage.UpdateCounter(k, v)
	}
	return nil
}

func (f *FileStorage) saveMetrics() error {
	tmpfile, err := os.CreateTemp("", "metrics-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	data := storageData{
		Gauges:   f.GetGauges(),
		Counters: f.GetCounters(),
	}

	if err := json.NewEncoder(tmpfile).Encode(data); err != nil {
		tmpfile.Close()
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpfile.Name(), f.path)
}

func (f *FileStorage) Save() error {
	if f.syncSave {
		return f.saveMetrics()
	}
	return nil
}

func (f *FileStorage) UpdateGauge(name string, value float64) {
	f.MemStorage.UpdateGauge(name, value)
	if f.syncSave {
		if err := f.saveMetrics(); err != nil {
			log.Printf("error saving metrics: %v", err)
		}
	}
}

func (f *FileStorage) UpdateCounter(name string, value int64) {
	f.MemStorage.UpdateCounter(name, value)
	if f.syncSave {
		if err := f.saveMetrics(); err != nil {
			log.Printf("error saving metrics: %v", err)
		}
	}
}

func (f *FileStorage) startAutoSave() {
	go func() {
		ticker := time.NewTicker(f.interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := f.saveMetrics(); err != nil {
				log.Printf("error saving metrics: %v", err)
			}
		}
	}()
}
