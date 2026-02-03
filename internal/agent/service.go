package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/idudko/go-musthave-metrics/internal/model"
)

type MetricsService struct {
	collector     *Collector
	sender        *Sender
	serverAddress string
	useBatch      bool
	rateLimit     int
	cryptoKey     string

	metricsChan chan []byte
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup

	workerPool *WorkerPool
}

func NewMetricsService(serverAddress, key string, useBatch bool, rateLimit int, cryptoKey string) *MetricsService {
	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsService{
		collector:     NewCollector(key),
		sender:        NewSender(key, cryptoKey),
		serverAddress: serverAddress,
		useBatch:      useBatch,
		rateLimit:     rateLimit,
		cryptoKey:     cryptoKey,
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

// Shutdown gracefully stops the service by:
// 1. Canceling the context to stop collection
// 2. Sending a final report of all metrics with a new context
// 3. Waiting for all pending tasks to complete
func (s *MetricsService) Shutdown() {
	log.Println("Agent shutdown: stopping metrics collection...")

	// First, cancel context to stop collectors
	s.cancel()

	// Wait for collectors to finish
	time.Sleep(100 * time.Millisecond)

	// Collect and send final metrics before shutdown with a new context
	log.Println("Agent shutdown: sending final metrics...")
	s.sendFinalMetrics()

	// Stop worker pool and wait for all tasks to complete
	log.Println("Agent shutdown: waiting for tasks to complete...")
	s.workerPool.Stop()

	// Wait for all goroutines to finish
	s.wg.Wait()

	log.Println("Agent shutdown: all operations completed")
}

// sendFinalMetrics sends all collected metrics with a new context
func (s *MetricsService) sendFinalMetrics() {
	// Create a new context with timeout for final metrics send
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gauges, counters := s.collector.GetMetrics()

	if s.useBatch {
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
			if err := s.sender.SendMetricsBatch(ctx, s.serverAddress, metrics); err != nil {
				log.Printf("Error sending final metrics batch: %v", err)
			}
		}
	} else {
		for name, value := range counters {
			m := model.Metrics{
				ID:    name,
				MType: model.Counter,
				Delta: &value,
			}
			if err := s.sender.SendMetricJSON(ctx, s.serverAddress, &m); err != nil {
				log.Printf("Error sending final counter metric %s: %v", name, err)
			}
		}

		for name, value := range gauges {
			m := model.Metrics{
				ID:    name,
				MType: model.Gauge,
				Value: &value,
			}
			if err := s.sender.SendMetricJSON(ctx, s.serverAddress, &m); err != nil {
				log.Printf("Error sending final gauge metric %s: %v", name, err)
			}
		}
	}
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
				return s.sender.SendMetricsBatch(ctx, s.serverAddress, metrics)
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
				return s.sender.SendMetricJSON(ctx, s.serverAddress, &m)
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
				return s.sender.SendMetricJSON(ctx, s.serverAddress, &m)
			})
		}
	}
}
