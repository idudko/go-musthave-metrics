package agent

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Collector struct {
	mu        sync.Mutex
	gauges    map[string]float64
	counters  map[string]int64
	pollCount int64
	key       string
}

func NewCollector(key string) *Collector {
	return &Collector{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
		key:      key,
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

func (c *Collector) CollectSystemMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	vmStat, err := mem.VirtualMemory()
	if err == nil {
		c.gauges["TotalMemory"] = float64(vmStat.Total)
		c.gauges["FreeMemory"] = float64(vmStat.Free)
	}

	// Используем true для получения per-CPU метрик
	cpuPercents, err := cpu.Percent(time.Second, true)
	if err == nil && len(cpuPercents) > 0 {
		for i, percent := range cpuPercents {
			c.gauges[fmt.Sprintf("CPUutilization%d", i+1)] = percent
		}
	}
}

func (c *Collector) GetMetrics() (map[string]float64, map[string]int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	gaugesCopy := make(map[string]float64, len(c.gauges))
	countersCopy := make(map[string]int64, len(c.counters))

	for k, v := range c.gauges {
		gaugesCopy[k] = v
	}

	for k, v := range c.counters {
		countersCopy[k] = v
	}

	return gaugesCopy, countersCopy
}
