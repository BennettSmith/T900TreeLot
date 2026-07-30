// Package postgres provides a shared PostgreSQL connection pool.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a connection pool used by adapters and entry points.
type DB struct {
	pool *pgxpool.Pool
}

// Open creates a pool and verifies connectivity.
func Open(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	db := &DB{pool: pool}
	if err := db.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return db, nil
}

// Ping verifies the database is reachable.
func (db *DB) Ping(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	return nil
}

// Close releases the pool.
func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

// Exec runs a statement that does not return rows.
func (db *DB) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, arguments...)
}

// QueryRow returns a single-row result.
func (db *DB) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	return db.pool.QueryRow(ctx, sql, arguments...)
}

// Query returns multiple rows.
func (db *DB) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, arguments...)
}

// Begin starts a transaction.
func (db *DB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// Pool exposes the underlying pool for advanced adapters.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}
