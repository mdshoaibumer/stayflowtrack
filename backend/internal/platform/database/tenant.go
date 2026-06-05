package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetTenantContext sets the app.current_tenant_id session variable on a connection.
// This enables RLS policies as a defense-in-depth layer.
func SetTenantContext(ctx context.Context, conn *pgxpool.Pool, tenantID uuid.UUID) error {
	// uuid.UUID.String() is guaranteed to produce a valid UUID format (no injection risk),
	// but we use parameterized set_config() for defense-in-depth.
	_, err := conn.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID.String())
	return err
}

// WithTenantTransaction begins a transaction with the tenant context set.
// RLS policies will be enforced for any queries within this transaction.
func WithTenantTransaction(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tenant tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set tenant context for RLS using parameterized set_config
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID.String())
	if err != nil {
		return fmt.Errorf("set tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
