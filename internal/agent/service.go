package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/idudko/go-musthave-metrics/internal/agent/grpc"
	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/netutil"
)

type MetricsService struct {
	collector     *Collector
	sender        *Sender
	serverAddress string
	grpcAddress   string
	useBatch      bool
	rateLimit     int
	cryptoKey     string

	metricsChan chan []byte
	grpcClient  *grpc.MetricsClient
	useGRPC     bool
	localIP     string
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup

	workerPool *WorkerPool
}

func NewMetricsService(serverAddress, grpcAddress, key string, useBatch bool, rateLimit int, cryptoKey string) *MetricsService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &MetricsService{
		collector:     NewCollector(key),
		sender:        NewSender(key, cryptoKey),
		serverAddress: serverAddress,
		grpcAddress:   grpcAddress,
		useBatch:      useBatch,
		rateLimit:     rateLimit,
		cryptoKey:     cryptoKey,
		metricsChan:   make(chan []byte, 100),
		ctx:           ctx,
		cancel:        cancel,
		workerPool:    NewWorkerPool(rateLimit),
	}

	// Инициализируем gRPC клиент, если указан адрес
	if grpcAddress != "" {
		client, err := grpc.NewMetricsClient(grpcAddress)
		if err != nil {
			log.Printf("Failed to create gRPC client: %v. Falling back to HTTP.", err)
		} else {
			service.grpcClient = client
			service.useGRPC = true
			log.Printf("gRPC client initialized for address: %s", grpcAddress)
		}
	}

	// Получаем локальный IP адрес для отправки в метаданных
	if service.useGRPC {
		localIP, err := netutil.GetLocalIP()
		if err != nil {
			log.Printf("Failed to get local IP: %v. Using empty IP.", err)
			service.localIP = ""
		} else {
			service.localIP = localIP
		}
	}

	return service
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

	// Close gRPC client if it was initialized
	if s.grpcClient != nil {
		s.grpcClient.Close()
	}

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

	// Convert gauges and counters to model.Metrics slice
	metrics := make([]model.Metrics, 0, len(counters)+len(gauges))

	for name, value := range counters {
		m := model.Metrics{
			ID:    name,
			MType: model.Counter,
			Delta: &value,
		}
		metrics = append(metrics, m)
	}
	for name, value := range gauges {
		m := model.Metrics{
			ID:    name,
			MType: model.Gauge,
			Value: &value,
		}
		metrics = append(metrics, m)
	}

	if s.useGRPC && len(metrics) > 0 {
		// Send via gRPC
		if err := s.grpcClient.UpdateMetrics(ctx, metrics, s.localIP); err != nil {
			log.Printf("Error sending final metrics via gRPC: %v. Falling back to HTTP.", err)
			// Fallback to HTTP if gRPC fails
			s.sendMetricsHTTP(ctx, metrics)
		}
	} else if len(metrics) > 0 {
		// Send via HTTP
		s.sendMetricsHTTP(ctx, metrics)
	}
}

// sendMetricsHTTP отправляет метрики через HTTP
func (s *MetricsService) sendMetricsHTTP(ctx context.Context, metrics []model.Metrics) error {
	if s.useBatch {
		// Convert to pointer slice for batch sending
		metricPtrs := make([]*model.Metrics, len(metrics))
		for i := range metrics {
			metricPtrs[i] = &metrics[i]
		}
		return s.sender.SendMetricsBatch(ctx, s.serverAddress, metricPtrs)
	} else {
		for _, m := range metrics {
			if err := s.sender.SendMetricJSON(ctx, s.serverAddress, &m); err != nil {
				return err
			}
		}
		return nil
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

	// Convert to model.Metrics slice
	metrics := make([]model.Metrics, 0, len(counters)+len(gauges))

	for name, value := range counters {
		m := model.Metrics{
			ID:    name,
			MType: model.Counter,
			Delta: &value,
		}
		metrics = append(metrics, m)
	}
	for name, value := range gauges {
		m := model.Metrics{
			ID:    name,
			MType: model.Gauge,
			Value: &value,
		}
		metrics = append(metrics, m)
	}

	if s.useGRPC && len(metrics) > 0 {
		// Send via gRPC (always batch)
		metricsCopy := make([]model.Metrics, len(metrics))
		copy(metricsCopy, metrics)
		s.workerPool.EnqueueTask(func(ctx context.Context) error {
			err := s.grpcClient.UpdateMetrics(ctx, metricsCopy, s.localIP)
			if err != nil {
				log.Printf("Error sending metrics via gRPC: %v. Falling back to HTTP.", err)
				// Fallback to HTTP if gRPC fails
				return s.sendMetricsHTTP(ctx, metricsCopy)
			}
			return nil
		})
	} else if len(metrics) > 0 {
		// Send via HTTP
		metricsCopy := make([]model.Metrics, len(metrics))
		copy(metricsCopy, metrics)
		s.workerPool.EnqueueTask(func(ctx context.Context) error {
			return s.sendMetricsHTTP(ctx, metricsCopy)
		})
	}
}
