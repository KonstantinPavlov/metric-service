package repository

import (
	"context"
	"fmt"

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
	}
	return nil
}

func (ps *PgStorage) Stop() {
	if ps.pool != nil {
		ps.log.Info("Stopping Postgres Storage...")
		ps.pool.Close()
	}
}

func (ps *PgStorage) Ping() error {
	if ps.pool == nil {
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
