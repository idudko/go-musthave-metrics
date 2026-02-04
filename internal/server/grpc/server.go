package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/idudko/go-musthave-metrics/internal/proto"
	"github.com/idudko/go-musthave-metrics/internal/repository"
	"github.com/rs/zerolog/log"
)

// MetricsServer реализует интерфейс gRPC сервиса Metrics
type MetricsServer struct {
	proto.UnimplementedMetricsServer
	storage repository.Storage
}

// UpdateMetrics обрабатывает запрос на обновление метрик
func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if len(req.Metrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no metrics provided")
	}

	// Обновляем метрики в хранилище
	for _, m := range req.Metrics {
		switch m.Type {
		case proto.Metric_GAUGE:
			if err := s.storage.UpdateGauge(ctx, m.Id, m.Value); err != nil {
				log.Error().Err(err).Str("id", m.Id).Float64("value", m.Value).Msg("failed to update gauge metric")
				return nil, status.Error(codes.Internal, "failed to update gauge metric")
			}
		case proto.Metric_COUNTER:
			if err := s.storage.UpdateCounter(ctx, m.Id, m.Delta); err != nil {
				log.Error().Err(err).Str("id", m.Id).Int64("delta", m.Delta).Msg("failed to update counter metric")
				return nil, status.Error(codes.Internal, "failed to update counter metric")
			}
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid metric type")
		}
	}

	return &proto.UpdateMetricsResponse{}, nil
}

// TrustedSubnetInterceptor проверяет, что IP из метаданных x-real-ip
// входит в доверенную подсеть
func TrustedSubnetInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Если доверенная подсеть не задана, просто пропускаем все запросы
		if trustedSubnet == "" {
			return handler(ctx, req)
		}

		// Парсим доверенную подсеть
		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			// Если CIDR невалидный, логируем предупреждение и пропускаем все запросы
			log.Warn().Err(err).Str("trusted_subnet", trustedSubnet).Msg("Invalid trusted subnet CIDR format, allowing all requests")
			return handler(ctx, req)
		}

		// Получаем IP из метаданных x-real-ip
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Warn().Msg("Failed to get metadata from context")
			return nil, status.Error(codes.PermissionDenied, "Failed to get metadata from context")
		}

		values := md.Get("x-real-ip")
		if len(values) == 0 {
			log.Warn().Msg("x-real-ip metadata is required but missing")
			return nil, status.Error(codes.PermissionDenied, "x-real-ip metadata is required")
		}

		realIP := values[0]
		// Парсим IP адрес
		ip := net.ParseIP(realIP)
		if ip == nil {
			log.Warn().Str("real_ip", realIP).Msg("Invalid IP address in x-real-ip metadata")
			return nil, status.Error(codes.PermissionDenied, "Invalid IP address in x-real-ip metadata")
		}

		// Проверяем, что IP входит в доверенную подсеть
		if !ipNet.Contains(ip) {
			log.Warn().Str("ip", ip.String()).Str("trusted_subnet", trustedSubnet).Msg("IP address is not in trusted subnet")
			return nil, status.Error(codes.PermissionDenied, "IP address is not in trusted subnet")
		}

		// IP в доверенной подсети, пропускаем запрос дальше
		log.Debug().Str("ip", ip.String()).Msg("IP is in trusted subnet, allowing request")
		return handler(ctx, req)
	}
}

// StartServer запускает gRPC сервер
func StartServer(ctx context.Context, address string, trustedSubnet string, storage repository.Storage) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	// Создаём интерцептор для проверки доверенной подсети
	interceptor := TrustedSubnetInterceptor(trustedSubnet)

	// Создаём gRPC сервер с интерцептором
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor))

	// Регистрируем сервис метрик
	metricsServer := &MetricsServer{
		storage: storage,
	}
	proto.RegisterMetricsServer(s, metricsServer)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Info().Str("address", address).Msg("Starting gRPC server")
		if err := s.Serve(lis); err != nil {
			log.Error().Err(err).Msg("Failed to start gRPC server")
		}
	}()

	// Обрабатываем graceful shutdown
	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down gRPC server gracefully...")
		s.GracefulStop()
	}()

	return s, nil
}
