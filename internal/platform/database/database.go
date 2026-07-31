// Package database provides Postgres connection and migration lifecycle helpers.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	// Register pgx as the database/sql driver used by Goose.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prabhavalabs/atlas/migrations"
	"github.com/pressly/goose/v3"
)

// Open creates a bounded connection pool and verifies the database is reachable.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("parse database configuration")
	}
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = "15s"
	poolConfig.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = "30s"

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Migrate applies all embedded Goose migrations to databaseURL.
func Migrate(ctx context.Context, databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}
	defer func() { _ = db.Close() }()

	goose.SetBaseFS(migrations.Files)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("configure migration dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// Ready verifies database reachability and schema compatibility.
func Ready(ctx context.Context, pool *pgxpool.Pool) error {
	var compatible bool
	err := pool.QueryRow(ctx, `
		SELECT application_version = 'foundation'
		FROM schema_metadata
		WHERE singleton = TRUE
	`).Scan(&compatible)
	if err != nil {
		return fmt.Errorf("read schema compatibility: %w", err)
	}
	if !compatible {
		return errors.New("database schema is incompatible with this Atlas build")
	}
	return nil
}
