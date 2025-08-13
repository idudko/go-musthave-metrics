package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type Collector struct {
	mu        sync.Mutex
	gauges    map[string]float64
	counters  map[string]int64
	pollCount int64
}

func NewCollector() *Collector {
	return &Collector{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (c *Collector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.gauges["Alloc"] = float64(memStats.Alloc)
	c.gauges["BuckHashSys"] = float64(memStats.BuckHashSys)
	c.gauges["Frees"] = float64(memStats.Frees)
	c.gauges["GCCPUFraction"] = memStats.GCCPUFraction
	c.gauges["GCSys"] = float64(memStats.GCSys)
	c.gauges["HeapAlloc"] = float64(memStats.HeapAlloc)
	c.gauges["HeapIdle"] = float64(memStats.HeapIdle)
	c.gauges["HeapInuse"] = float64(memStats.HeapInuse)
	c.gauges["HeapObjects"] = float64(memStats.HeapObjects)
	c.gauges["HeapReleased"] = float64(memStats.HeapReleased)
	c.gauges["HeapSys"] = float64(memStats.HeapSys)
	c.gauges["LastGC"] = float64(memStats.LastGC)
	c.gauges["Lookups"] = float64(memStats.Lookups)
	c.gauges["MCacheInuse"] = float64(memStats.MCacheInuse)
	c.gauges["MCacheSys"] = float64(memStats.MCacheSys)
	c.gauges["MSpanInuse"] = float64(memStats.MSpanInuse)
	c.gauges["MSpanSys"] = float64(memStats.MSpanSys)
	c.gauges["Mallocs"] = float64(memStats.Mallocs)
	c.gauges["NextGC"] = float64(memStats.NextGC)
	c.gauges["NumForcedGC"] = float64(memStats.NumForcedGC)
	c.gauges["NumGC"] = float64(memStats.NumGC)
	c.gauges["OtherSys"] = float64(memStats.OtherSys)
	c.gauges["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	c.gauges["StackInuse"] = float64(memStats.StackInuse)
	c.gauges["StackSys"] = float64(memStats.StackSys)
	c.gauges["Sys"] = float64(memStats.Sys)
	c.gauges["TotalAlloc"] = float64(memStats.TotalAlloc)

	c.gauges["RandomValue"] = rand.Float64()

	c.pollCount++
	c.counters["PollCount"] = c.pollCount
}

func (c *Collector) Report(serverAddress string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for name, value := range c.gauges {
		url := fmt.Sprintf("%s/update/gauge/%s/%f", serverAddress, name, value)
		if err := sendMetric(url); err != nil {
			return err
		}
	}
	for name, value := range c.counters {
		url := fmt.Sprintf("%s/update/counter/%s/%d", serverAddress, name, value)
		if err := sendMetric(url); err != nil {
			return err
		}
	}
	return nil
}

func sendMetric(url string) error {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
