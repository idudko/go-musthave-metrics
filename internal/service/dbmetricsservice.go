package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBMetricsService struct {
	pool *pgxpool.Pool
}

func NewDBMetricsService(pool *pgxpool.Pool) *DBMetricsService {
	return &DBMetricsService{pool: pool}
}

func (s *DBMetricsService) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
