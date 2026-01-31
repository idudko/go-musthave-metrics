package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/idudko/go-musthave-metrics/pkg/hash"
)

type Sender struct {
	key string
}

func NewSender(key string) *Sender {
	return &Sender{
		key: key,
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

	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	if _, err := gw.Write(data); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}

	compressed := b.Bytes()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if s.key != "" {
		hashValue := hash.ComputeHash(compressed, s.key)
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

	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	if _, err := gw.Write(data); err != nil {
		return fmt.Errorf("failed to write data to gzip writer: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}
	compressed := b.Bytes()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if s.key != "" {
		hashValue := hash.ComputeHash(compressed, s.key)
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
