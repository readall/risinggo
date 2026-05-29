package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/readall/risinggo/internal/config"
)

type Pool struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

func NewPool(cfg *config.Config) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnIdleTime = 10 * time.Minute
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.HealthCheckPeriod = 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	return &Pool{pool: pool, cfg: cfg}, nil
}

func (p *Pool) Close() {
	p.pool.Close()
}

func (p *Pool) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	return 0, fmt.Errorf("mutations not allowed - use Query instead")
}

type PgxRows struct {
	pgx.Rows
}

func (p *Pool) Query(ctx context.Context, sql string, args ...any) (*PgxRows, error) {
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return &PgxRows{Rows: rows}, nil
}

func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *Pool) Stats() *pgxpool.Stat {
	return p.pool.Stat()
}

// Ping performs a health check against the database.
func (p *Pool) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}
