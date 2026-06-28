package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stayflow/stayflow-track/internal/config"
)

const (
	maxRetries    = 5
	retryBaseWait = 2 * time.Second
)

// Connect establishes a database connection pool with retry logic.
// It will retry up to 5 times with exponential backoff if the database
// is not immediately available (common during container startup).
func Connect(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.MaxLifetime

	var pool *pgxpool.Pool
	for attempt := 0; attempt < maxRetries; attempt++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
		if err != nil {
			waitTime := retryBaseWait * time.Duration(1<<uint(attempt))
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during db connect retry: %w", ctx.Err())
			case <-time.After(waitTime):
				continue
			}
		}

		if pingErr := pool.Ping(ctx); pingErr != nil {
			pool.Close()
			waitTime := retryBaseWait * time.Duration(1<<uint(attempt))
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during db ping retry: %w", ctx.Err())
			case <-time.After(waitTime):
				continue
			}
		}

		return pool, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// ConnectAppRole establishes a connection pool using the restricted stayflow_app role.
// This role is subject to RLS policies and should be used for all tenant-scoped queries
// in production. The owner pool (Connect) is used for migrations and admin operations.
func ConnectAppRole(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	appCfg := cfg
	if cfg.AppUser != "" {
		appCfg.User = cfg.AppUser
		appCfg.Password = cfg.AppPassword
	}
	return Connect(ctx, appCfg)
}
