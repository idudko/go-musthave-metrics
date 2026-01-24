package agent

import (
	"context"
	"sync"
	"time"

	"github.com/idudko/go-musthave-metrics/internal/model"
)

type MetricsService struct {
	collector     *Collector
	serverAddress string
	useBatch      bool
	rateLimit     int

	metricsChan chan []byte
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup

	workerPool *WorkerPool
}

func NewMetricsService(serverAddress, key string, useBatch bool, rateLimit int) *MetricsService {
	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsService{
		collector:     NewCollector(key),
		serverAddress: serverAddress,
		useBatch:      useBatch,
		rateLimit:     rateLimit,
		metricsChan:   make(chan []byte, 100),
		ctx:           ctx,
		cancel:        cancel,
		workerPool:    NewWorkerPool(rateLimit),
	}
}

func (s *MetricsService) Start(pollInterval, reportInterval int) {
	s.workerPool.Start(s.ctx)

	s.wg.Add(1)
	go s.collectRuntimeMetrics(pollInterval)

	s.wg.Add(1)
	go s.collectSystemMetrics(pollInterval)

	s.wg.Add(1)
	go s.sendMetrics(reportInterval)
}

func (s *MetricsService) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *MetricsService) collectRuntimeMetrics(pollInterval int) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.collector.Collect()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *MetricsService) collectSystemMetrics(pollInterval int) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.collector.CollectSystemMetrics()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *MetricsService) sendMetrics(reportInterval int) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.enqueueMetricsForSending()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *MetricsService) enqueueMetricsForSending() {
	gauges, counters := s.collector.GetMetrics()

	if s.useBatch {
		s.workerPool.EnqueueTask(func(ctx context.Context) error {
			metrics := make([]*model.Metrics, 0, len(counters)+len(gauges))

			for name, value := range counters {
				m := model.Metrics{
					ID:    name,
					MType: model.Counter,
					Delta: &value,
				}
				metrics = append(metrics, &m)
			}
			for name, value := range gauges {
				m := model.Metrics{
					ID:    name,
					MType: model.Gauge,
					Value: &value,
				}
				metrics = append(metrics, &m)
			}

			if len(metrics) > 0 {
				return s.collector.sendMetricsBatch(ctx, s.serverAddress, metrics)
			}
			return nil
		})
	} else {
		for name, value := range counters {
			name := name
			value := value
			s.workerPool.EnqueueTask(func(ctx context.Context) error {
				m := model.Metrics{
					ID:    name,
					MType: model.Counter,
					Delta: &value,
				}
				return s.collector.sendMetricJSON(ctx, s.serverAddress, &m)
			})
		}

		for name, value := range gauges {
			name := name
			value := value
			s.workerPool.EnqueueTask(func(ctx context.Context) error {
				m := model.Metrics{
					ID:    name,
					MType: model.Gauge,
					Value: &value,
				}
				return s.collector.sendMetricJSON(ctx, s.serverAddress, &m)
			})
		}
	}
}
