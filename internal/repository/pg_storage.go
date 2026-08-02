package repository

import (
	"context"
	"database/sql"

	"errors"
	"fmt"

	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

type PgStorage struct {
	ctx        context.Context
	connString string
	log        *zap.Logger
	pool       *pgxpool.Pool
}

type StoragePinger interface {
	Ping() error
}

func NewPgStorage(ctx context.Context, connString string, log *zap.Logger) *PgStorage {
	return &PgStorage{
		ctx:        ctx,
		connString: connString,
		log:        log,
	}
}

func (ps *PgStorage) Start() error {
	if ps.connString == "" {
		log.Warn("Connection string is empty! Postgres Storage not started!")
	} else {
		pool, err := pgxpool.New(ps.ctx, ps.connString)
		if err != nil {
			return err
		}
		ps.log.Info("Postgres Storage started!")
		ps.pool = pool
		// Start migartions
		sourceURL := "file://migrations/."
		migrationURL := "pgx5" + ps.connString[8:]
		m, err := migrate.New(sourceURL, migrationURL)
		if err != nil {
			ps.log.Error("Failed to initialize migrate!", zap.Error(err))
			return err
		}
		defer m.Close()
		ps.log.Info("Running database migrations from disk...")
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				ps.log.Info("Database is up to date. No migrations applied.")
				return nil
			}
			ps.log.Error("Failed to run up migrations: %v", zap.Error(err))
			return err
		}
		ps.log.Info("Migrations applied successfully!")
	}
	return nil
}

func (ps *PgStorage) Stop() {
	if ps.started() {
		ps.log.Info("Stopping Postgres Storage...")
		ps.pool.Close()
	}
}

func (ps *PgStorage) Ping() error {
	if !ps.started() {
		return fmt.Errorf("PgStorage not started!")
	}
	err := ps.pool.Ping(ps.ctx)
	if err != nil {
		log.Error("Failed to ping db!", zap.Error(err))
		return err
	}
	log.Info("Ping ok!")
	return nil
}
func (ps *PgStorage) started() bool {
	return ps.pool != nil
}

func (ps *PgStorage) GetNames(metricType string) ([]string, error) {
	if !ps.started() {
		return nil, fmt.Errorf("Storage not started!")
	}
	var query string
	switch metricType {
	case model.Counter:
		query = "SELECT name FROM public.counters"
	case model.Gauge:
		query = "SELECT name FROM public.gauges"
	default:
		return make([]string, 0), fmt.Errorf("Unknown metricType %q", metricType)
	}

	rows, err := ps.pool.Query(ps.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Unable to execute query: %w", err)
	}
	defer rows.Close()

	var names []string

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("Failed to scan row: %w", err)
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error during rows iteration: %w", err)
	}

	return names, nil
}

func (ps *PgStorage) GetCounter(name string) (*MetricData, error) {
	if !ps.started() {
		return nil, fmt.Errorf("Storage not started!")
	}
	query := `SELECT delta FROM public.counters WHERE name = $1;`
	var delta sql.NullInt64

	err := ps.pool.QueryRow(ps.ctx, query, name).Scan(&delta)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to query counter: %w", err)
	}
	var finalValue interface{}
	if delta.Valid {
		finalValue = delta.Int64
	} else {
		finalValue = nil
	}

	return &MetricData{
		Name:  name,
		Value: finalValue,
	}, nil
}

func (ps *PgStorage) GetGauge(name string) (*MetricData, error) {
	if !ps.started() {
		return nil, fmt.Errorf("Storage not started!")
	}
	query := `SELECT value FROM public.gauges WHERE name = $1;`

	var value sql.NullFloat64

	err := ps.pool.QueryRow(ps.ctx, query, name).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to query gauge: %w", err)
	}

	var finalValue interface{}
	if value.Valid {
		finalValue = value.Float64
	} else {
		finalValue = nil
	}

	return &MetricData{
		Name:  name,
		Value: finalValue,
	}, nil
}

func (ps *PgStorage) SaveCounter(name string, value int64) error {
	if !ps.started() {
		return fmt.Errorf("Storage not started!")
	}
	query := `
		INSERT INTO public.counters (name, delta) 
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET delta = counters.delta + EXCLUDED.delta;
	`
	_, err := ps.pool.Exec(ps.ctx, query, name, value)
	if err != nil {
		return fmt.Errorf("Failed to save counter %q: %w", name, err)
	}
	return nil
}

func (ps *PgStorage) SaveGauge(name string, value float64) error {
	if !ps.started() {
		return fmt.Errorf("Storage not started!")
	}
	query := `
		INSERT INTO public.gauges (name, value) 
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET value = EXCLUDED.value;
	`

	_, err := ps.pool.Exec(ps.ctx, query, name, value)
	if err != nil {
		return fmt.Errorf("Failed to save gauge %q: %w", name, err)
	}

	return nil
}

func (ps *PgStorage) SaveCounters(counters []MetricData) error {
	if !ps.started() {
		return fmt.Errorf("Storage not started!")
	}

	if len(counters) == 0 {
		return nil
	}

	query := `
		INSERT INTO public.counters (name, delta) 
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET delta = counters.delta + EXCLUDED.delta;
	`
	batch := &pgx.Batch{}
	for _, c := range counters {
		batch.Queue(query, c.Name, c.Value)
	}
	br := ps.pool.SendBatch(ps.ctx, batch)
	defer br.Close()

	// Проверяем статус выполнения каждого элемента пакета
	for i := 0; i < len(counters); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("Failed to execute batch counter item %d: %w", i, err)
		}
	}

	return nil
}

func (ps *PgStorage) SaveGauges(gauges []MetricData) error {
	if !ps.started() {
		return fmt.Errorf("Storage not started!")
	}
	if len(gauges) == 0 {
		return nil
	}
	query := `
		INSERT INTO public.gauges (name, value) 
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET value = EXCLUDED.value;
	`
	batch := &pgx.Batch{}
	for _, g := range gauges {
		batch.Queue(query, g.Name, g.Value)
	}

	br := ps.pool.SendBatch(ps.ctx, batch)
	defer br.Close()

	for i := 0; i < len(gauges); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("Failed to execute batch item %d: %w", i, err)
		}
	}

	return nil
}
