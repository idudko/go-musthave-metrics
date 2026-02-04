package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/pkg/crypto"
	"github.com/idudko/go-musthave-metrics/pkg/hash"
	"github.com/rs/zerolog/log"
)

type Sender struct {
	key       string
	cryptoKey *rsa.PublicKey
}

func NewSender(key string, cryptoKeyPath string) *Sender {
	var cryptoKey *rsa.PublicKey
	if cryptoKeyPath != "" {
		pubKey, err := crypto.LoadPublicKey(cryptoKeyPath)
		if err != nil {
			// Log error but continue - encryption will be disabled
			log.Printf("Failed to load public key: %v. Encryption will be disabled.", err)
		} else {
			cryptoKey = pubKey
		}
	}

	return &Sender{
		key:       key,
		cryptoKey: cryptoKey,
	}
}

func (s *Sender) SendMetricJSON(ctx context.Context, serverAddress string, m *model.Metrics) error {
	url := fmt.Sprintf("http://%s/update", serverAddress)
	retryIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	err := s.doSendMetricJSON(ctx, url, m)
	if err == nil {
		return nil
	}

	for _, interval := range retryIntervals {
		select {
		case <-time.After(interval):
			err = s.doSendMetricJSON(ctx, url, m)
			if err == nil {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

func (s *Sender) doSendMetricJSON(ctx context.Context, url string, m *model.Metrics) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("failed to get local IP: %w", err)
	}

	var b bytes.Buffer
	var requestBody []byte
	var contentEncoding string

	if s.cryptoKey != nil {
		// Encrypt with RSA public key
		encryptedData, err := crypto.Encrypt(data, s.cryptoKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
		requestBody = encryptedData
		b.Write(encryptedData)
		contentEncoding = "encrypt"
	} else {
		// Compress with gzip
		gw := gzip.NewWriter(&b)
		if _, err := gw.Write(data); err != nil {
			return err
		}
		if err := gw.Close(); err != nil {
			return err
		}
		requestBody = b.Bytes()
		contentEncoding = "gzip"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", contentEncoding)
	req.Header.Set("X-Real-IP", localIP)

	if s.key != "" && s.cryptoKey == nil {
		hashValue := hash.ComputeHash(requestBody, s.key)
		req.Header.Set("HashSHA256", hashValue)
	}

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

func (s *Sender) SendMetricsBatch(ctx context.Context, serverAddress string, metrics []*model.Metrics) error {
	url := fmt.Sprintf("http://%s/updates", serverAddress)
	retryIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	err := s.doSendMetricsBatch(ctx, url, metrics)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "400") {
		return nil
	}

	for _, interval := range retryIntervals {
		select {
		case <-time.After(interval):
			err = s.doSendMetricsBatch(ctx, url, metrics)
			if err == nil {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

func (s *Sender) doSendMetricsBatch(ctx context.Context, url string, metrics []*model.Metrics) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("failed to get local IP: %w", err)
	}

	var b bytes.Buffer
	var requestBody []byte
	var contentEncoding string

	if s.cryptoKey != nil {
		// Encrypt with RSA public key
		encryptedData, err := crypto.Encrypt(data, s.cryptoKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
		requestBody = encryptedData
		b.Write(encryptedData)
		contentEncoding = "encrypt"
	} else {
		// Compress with gzip
		gw := gzip.NewWriter(&b)
		if _, err := gw.Write(data); err != nil {
			return fmt.Errorf("failed to write data to gzip writer: %w", err)
		}
		if err := gw.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
		requestBody = b.Bytes()
		contentEncoding = "gzip"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", contentEncoding)
	req.Header.Set("X-Real-IP", localIP)

	if s.key != "" && s.cryptoKey == nil {
		hashValue := hash.ComputeHash(requestBody, s.key)
		req.Header.Set("HashSHA256", hashValue)
	}

	client := &http.Client{Timeout: 30 * time.Second}
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

// getLocalIP возвращает локальный IP адрес хоста
func getLocalIP() (string, error) {
	// Получаем все сетевые интерфейсы
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, addr := range addrs {
		// Пропускаем локальный loopback интерфейс
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// Предпочитаем IPv4 адреса
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid local IP address found")
}
