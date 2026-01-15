package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/idudko/go-musthave-metrics/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbBatchSize = 500
)

type DBStorage struct {
	pool *pgxpool.Pool
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DBStorage{pool: pool}

	if err := db.runMigrations(dsn); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func (d *DBStorage) runMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("Database is up to date")
	} else {
		log.Println("Database migrations completed")
	}

	return nil
}

func (d *DBStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	query := `
		INSERT INTO gauges (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = $2
	`
	_, err := d.pool.Exec(ctx, query, name, value)
	return err
}

func (d *DBStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	query := `
		INSERT INTO counters (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = counters.value + $2
	`
	_, err := d.pool.Exec(ctx, query, name, value)
	return err
}

func (d *DBStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	query := `
		SELECT name, value FROM gauges
	`
	rows, err := d.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		result[name] = value
	}
	return result, rows.Err()
}

func (d *DBStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT name, value FROM counters
	`
	rows, err := d.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		result[name] = value
	}
	return result, rows.Err()
}

func (d *DBStorage) Save(ctx context.Context) error {
	return nil
}

func (d *DBStorage) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

func (d *DBStorage) Close() {
	d.pool.Close()
}

func (d *DBStorage) UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var gaugeBatch [][]any
	var counterBatch [][]any

	for _, metric := range metrics {
		switch metric.MType {
		case model.Gauge:
			if metric.Value != nil {
				gaugeBatch = append(gaugeBatch, []any{metric.ID, metric.Value})
			}
		case model.Counter:
			if metric.Delta != nil {
				counterBatch = append(counterBatch, []any{metric.ID, metric.Delta})
			}
		}
	}

	if len(gaugeBatch) > 0 {
		for _, row := range gaugeBatch {
			query := `
				INSERT INTO gauges (name, value)
				VALUES ($1, $2)
			 	ON CONFLICT (name) DO UPDATE SET value = $2
				`
			if _, err := tx.Exec(ctx, query, row[0], row[1]); err != nil {
				return fmt.Errorf("failed to update gauge: %w", err)
			}
		}
	}

	if len(counterBatch) > 0 {
		for _, row := range counterBatch {
			query := `
				INSERT INTO counters (name, value)
				VALUES ($1, $2)
			 	ON CONFLICT (name) DO UPDATE SET value = counters.value + $2
				`
			if _, err := tx.Exec(ctx, query, row[0], row[1]); err != nil {
				return fmt.Errorf("failed to update counter: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
