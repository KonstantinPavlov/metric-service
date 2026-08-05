package repository

import (
	"context"
	"database/sql"
	"time"

	"errors"
	"fmt"

	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/KonstantinPavlov/metric-service/internal/pgerrors"
	"github.com/golang-migrate/migrate/v4"
	migratePgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type PgStorage struct {
	connString string
	log        *zap.Logger
	pool       *pgxpool.Pool
}

func NewPgStorage(connString string, log *zap.Logger) *PgStorage {
	return &PgStorage{
		connString: connString,
		log:        log,
	}
}

func executeWithRetry(ctx context.Context, logger *zap.Logger, operation func() error) error {
	retryDelay := 1 * time.Second
	maxAttempts := 4

	classifier := pgerrors.NewPostgresErrorClassifier()

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := operation()
		if err == nil {
			return nil
		}
		classification := classifier.Classify(err)

		if attempt == maxAttempts || classification == pgerrors.NonRetriable {
			return err
		}

		currentDelay := retryDelay
		logger.Warn("Database transport error occurred, retrying...",
			zap.Int("attempt", attempt),
			zap.Duration("wait_time", currentDelay),
			zap.Error(err),
		)

		select {
		case <-time.After(currentDelay):
			retryDelay += 2 * time.Second
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (ps *PgStorage) Start(ctx context.Context) error {
	if ps.connString == "" {
		ps.log.Warn("Connection string is empty! Postgres Storage not started!")
	} else {
		pool, err := pgxpool.New(ctx, ps.connString)
		if err != nil {
			return err
		}
		ps.log.Info("Postgres Storage started!")
		ps.pool = pool
		// Start migartions
		err = executeWithRetry(ctx, ps.log, func() error {
			sourceURL := "file://migrations/."
			db := stdlib.OpenDBFromPool(ps.pool)
			driver, err := migratePgx.WithInstance(db, &migratePgx.Config{})
			if err != nil {
				return fmt.Errorf("Аailed to create migrate driver: %w", err)
			}

			m, err := migrate.NewWithDatabaseInstance(sourceURL, "pgx5", driver)
			if err != nil {
				return fmt.Errorf("Аailed to initialize migrate: %w", err)
			}
			defer m.Close()
			ps.log.Info("Running database migrations from disk...")
			if err := m.Up(); err != nil {
				if errors.Is(err, migrate.ErrNoChange) {
					return nil
				}
				return err
			}
			return nil
		})

		if err != nil {
			ps.log.Error("Failed to apply migrations after retries", zap.Error(err))
			return err
		}

		ps.log.Info("Migrations applied successfully!")
		return nil
	}
	return nil
}

func (ps *PgStorage) Stop() {
	if ps.started() {
		ps.log.Info("Stopping Postgres Storage...")
		ps.pool.Close()
	}
}

func (ps *PgStorage) Ping(ctx context.Context) error {
	if !ps.started() {
		return fmt.Errorf("PgStorage not started!")
	}
	return executeWithRetry(ctx, ps.log, func() error {
		err := ps.pool.Ping(ctx)
		if err != nil {
			ps.log.Error("Failed to ping db!", zap.Error(err))
			return err
		}
		ps.log.Info("Ping ok!")
		return nil
	})
}

func (ps *PgStorage) started() bool {
	return ps.pool != nil
}

func (ps *PgStorage) GetNames(ctx context.Context, metricType string) ([]string, error) {
	if !ps.started() {
		return nil, fmt.Errorf("Storage not started!")
	}

	var names []string
	err := executeWithRetry(ctx, ps.log, func() error {
		var query string
		switch metricType {
		case model.Counter:
			query = "SELECT name FROM public.counters"
		case model.Gauge:
			query = "SELECT name FROM public.gauges"
		default:
			return fmt.Errorf("Unknown metricType %q", metricType)
		}

		rows, err := ps.pool.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		names = nil // Сбрасываем срез на случай повторной попытки
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return fmt.Errorf("Failed to scan row: %w", err)
			}
			names = append(names, name)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to execute query: %w", err)
	}
	return names, nil
}

func (ps *PgStorage) GetCounter(ctx context.Context, name string) (*MetricData, error) {
	if !ps.started() {
		return nil, fmt.Errorf("Storage not started!")
	}

	var result *MetricData
	err := executeWithRetry(ctx, ps.log, func() error {
		query := `SELECT delta FROM public.counters WHERE name = $1;`
		var delta sql.NullInt64

		err := ps.pool.QueryRow(ctx, query, name).Scan(&delta)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				result = nil
				return nil
			}
			return err
		}

		var finalValue interface{}
		if delta.Valid {
			finalValue = delta.Int64
		} else {
			finalValue = nil
		}

		result = &MetricData{Name: name, Value: finalValue}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to query counter: %w", err)
	}
	return result, nil
}

func (ps *PgStorage) GetGauge(ctx context.Context, name string) (*MetricData, error) {
	if !ps.started() {
		return nil, fmt.Errorf("Storage not started!")
	}

	var result *MetricData
	err := executeWithRetry(ctx, ps.log, func() error {
		query := `SELECT value FROM public.gauges WHERE name = $1;`
		var value sql.NullFloat64

		err := ps.pool.QueryRow(ctx, query, name).Scan(&value)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				result = nil
				return nil
			}
			return err
		}

		var finalValue interface{}
		if value.Valid {
			finalValue = value.Float64
		} else {
			finalValue = nil
		}

		result = &MetricData{Name: name, Value: finalValue}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to query gauge: %w", err)
	}
	return result, nil
}

func (ps *PgStorage) SaveCounter(ctx context.Context, name string, value int64) error {
	if !ps.started() {
		return fmt.Errorf("Storage not started!")
	}
	return executeWithRetry(ctx, ps.log, func() error {
		query := `
			INSERT INTO public.counters (name, delta) 
			VALUES ($1, $2)
			ON CONFLICT (name) 
			DO UPDATE SET delta = counters.delta + EXCLUDED.delta;
		`
		_, err := ps.pool.Exec(ctx, query, name, value)
		if err != nil {
			return fmt.Errorf("Failed to save counter %q: %w", name, err)
		}
		return nil
	})
}

func (ps *PgStorage) SaveGauge(ctx context.Context, name string, value float64) error {
	if !ps.started() {
		return fmt.Errorf("Storage not started!")
	}
	return executeWithRetry(ctx, ps.log, func() error {
		query := `
			INSERT INTO public.gauges (name, value) 
			VALUES ($1, $2)
			ON CONFLICT (name) 
			DO UPDATE SET value = EXCLUDED.value;
		`
		_, err := ps.pool.Exec(ctx, query, name, value)
		if err != nil {
			return fmt.Errorf("Failed to save gauge %q: %w", name, err)
		}
		return nil
	})
}

func (ps *PgStorage) SaveMetrics(ctx context.Context, counters []MetricData, gauges []MetricData) error {
	if !ps.started() {
		return fmt.Errorf("Storage not started!")
	}

	return executeWithRetry(ctx, ps.log, func() error {
		tx, err := ps.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		if len(counters) > 0 {
			counterQuery := `
				INSERT INTO public.counters (name, delta) 
				VALUES ($1, $2)
				ON CONFLICT (name) 
				DO UPDATE SET delta = counters.delta + EXCLUDED.delta;
			`
			for _, c := range counters {
				val, ok := c.Value.(int64)
				if !ok {
					return fmt.Errorf("invalid counter value type for %s", c.Name)
				}
				_, err := tx.Exec(ctx, counterQuery, c.Name, val)
				if err != nil {
					return fmt.Errorf("failed to exec counter %s: %w", c.Name, err)
				}
			}
		}

		if len(gauges) > 0 {
			gaugeQuery := `
				INSERT INTO public.gauges (name, value) 
				VALUES ($1, $2)
				ON CONFLICT (name) 
				DO UPDATE SET value = EXCLUDED.value;
			`
			for _, g := range gauges {
				val, ok := g.Value.(float64)
				if !ok {
					return fmt.Errorf("Invalid gauge value type for %s", g.Name)
				}
				_, err := tx.Exec(ctx, gaugeQuery, g.Name, val)
				if err != nil {
					return fmt.Errorf("failed to exec gauge %s: %w", g.Name, err)
				}
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("Failed to commit transaction: %w", err)
		}

		return nil
	})
}
