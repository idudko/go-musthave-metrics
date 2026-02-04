package grpc

import (
	"context"
	"errors"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/internal/proto"
	"github.com/rs/zerolog/log"
)

// MetricsClient представляет gRPC клиент для отправки метрик
type MetricsClient struct {
	client     proto.MetricsClient
	conn       *grpc.ClientConn
	serverAddr string
}

// NewMetricsClient создаёт новый gRPC клиент для работы с метриками
func NewMetricsClient(serverAddr string) (*MetricsClient, error) {
	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := &MetricsClient{
		client:     proto.NewMetricsClient(conn),
		conn:       conn,
		serverAddr: serverAddr,
	}

	return client, nil
}

// Close закрывает соединение с сервером
func (c *MetricsClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// UpdateMetrics отправляет метрики на сервер через gRPC
// metrics - список метрик для отправки
// clientIP - IP-адрес клиента, который будет добавлен в метаданные
func (c *MetricsClient) UpdateMetrics(ctx context.Context, metrics []model.Metrics, clientIP string) error {
	// Создаём запрос с метриками
	req := &proto.UpdateMetricsRequest{
		Metrics: make([]*proto.Metric, 0, len(metrics)),
	}

	// Конвертируем метрики в protobuf формат
	for _, m := range metrics {
		protoMetric := &proto.Metric{
			Id:   m.ID,
			Type: convertStringToProtoMType(m.MType),
		}

		switch m.MType {
		case model.Gauge:
			if m.Value != nil {
				protoMetric.Value = *m.Value
			}
		case model.Counter:
			if m.Delta != nil {
				protoMetric.Delta = *m.Delta
			}
		}

		req.Metrics = append(req.Metrics, protoMetric)
	}

	// Добавляем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Создаём метаданные с IP-адресом клиента
	md := &metadata.MD{
		"x-real-ip": []string{clientIP},
	}
	ctx = metadata.NewOutgoingContext(ctx, *md)

	// Отправляем запрос на сервер
	_, err := c.client.UpdateMetrics(ctx, req)
	if err != nil {
		log.Error().Err(err).Str("server_addr", c.serverAddr).Int("metrics_count", len(metrics)).Msg("Failed to send metrics via gRPC")
		return err
	}

	log.Debug().Int("metrics_count", len(metrics)).Str("client_ip", clientIP).Msg("Successfully sent metrics via gRPC")
	return nil
}

// convertStringToProtoMType конвертирует строковый тип метрики в protobuf тип
func convertStringToProtoMType(mType string) proto.Metric_MType {
	switch mType {
	case model.Gauge:
		return proto.Metric_GAUGE
	case model.Counter:
		return proto.Metric_COUNTER
	default:
		return proto.Metric_MType(0)
	}
}

// GetLocalIP возвращает локальный IP-адрес клиента
func GetLocalIP() (string, error) {
	// Получаем все сетевые интерфейсы
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// Проходим по всем интерфейсам и ищем подходящий IP
	for _, iface := range interfaces {
		// Пропускаем loopback и неактивные интерфейсы
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addresses {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Пропускаем IPv6 и loopback адреса
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			return ip.String(), nil
		}
	}

	// Если не нашли подходящий IP, возвращаем ошибку
	return "", errors.New("no network interface found")
}
