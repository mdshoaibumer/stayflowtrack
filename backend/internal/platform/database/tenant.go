package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tenantContextKey struct{}
type platformAdminContextKey struct{}

// WithTenantID attaches the authenticated tenant to the request context.
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenantID)
}

// TenantIDFromContext returns the authenticated tenant from the context.
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(tenantContextKey{}).(uuid.UUID)
	return tenantID, ok && tenantID != uuid.Nil
}

// WithPlatformAdmin marks a request as authorized for cross-tenant platform data.
func WithPlatformAdmin(ctx context.Context) context.Context {
	return context.WithValue(ctx, platformAdminContextKey{}, true)
}

func isPlatformAdmin(ctx context.Context) bool {
	isAdmin, _ := ctx.Value(platformAdminContextKey{}).(bool)
	return isAdmin
}

// TenantPool wraps pgxpool.Pool and applies RLS tenant context automatically.
type TenantPool struct {
	pool *pgxpool.Pool
}

func NewTenantPool(pool *pgxpool.Pool) *TenantPool {
	return &TenantPool{pool: pool}
}

// SetTenantContext sets tenant session variables on a connection.
// This enables RLS policies as a defense-in-depth layer.
func SetTenantContext(ctx context.Context, conn *pgxpool.Pool, tenantID uuid.UUID) error {
	_, err := conn.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true), set_config('app.current_tenant_id', $1, true)", tenantID.String())
	return err
}

// WithTenantTransaction begins a transaction with the tenant context set.
// RLS policies will be enforced for any queries within this transaction.
func WithTenantTransaction(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tenant tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := setTenantLocal(ctx, tx, tenantID, false); err != nil {
		return fmt.Errorf("set tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *TenantPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	tenantID, ok := TenantIDFromContext(ctx)
	if !ok {
		return p.pool.QueryRow(ctx, sql, args...)
	}

	tx, err := p.beginTenantTx(ctx, tenantID)
	if err != nil {
		return errRow{err: err}
	}

	return &tenantRow{
		ctx: ctx,
		tx:  tx,
		row: tx.QueryRow(ctx, sql, args...),
	}
}

func (p *TenantPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	tenantID, ok := TenantIDFromContext(ctx)
	if !ok {
		return p.pool.Query(ctx, sql, args...)
	}

	tx, err := p.beginTenantTx(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}

	return &tenantRows{ctx: ctx, tx: tx, rows: rows}, nil
}

func (p *TenantPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tenantID, ok := TenantIDFromContext(ctx)
	if !ok {
		return p.pool.Exec(ctx, sql, args...)
	}

	tx, err := p.beginTenantTx(ctx, tenantID)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return tag, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tag, err
	}
	return tag, nil
}

func (p *TenantPool) Begin(ctx context.Context) (pgx.Tx, error) {
	tenantID, ok := TenantIDFromContext(ctx)
	if !ok {
		return p.pool.Begin(ctx)
	}
	return p.beginTenantTx(ctx, tenantID)
}

func (p *TenantPool) beginTenantTx(ctx context.Context, tenantID uuid.UUID) (pgx.Tx, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tenant tx: %w", err)
	}
	if err := setTenantLocal(ctx, tx, tenantID, isPlatformAdmin(ctx)); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	return tx, nil
}

func setTenantLocal(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, platformAdmin bool) error {
	_, err := tx.Exec(ctx,
		`SELECT set_config('app.current_tenant', $1, true),
		        set_config('app.current_tenant_id', $1, true),
		        set_config('app.is_platform_admin', $2, true)`,
		tenantID.String(), fmt.Sprintf("%t", platformAdmin),
	)
	return err
}

type tenantRow struct {
	ctx context.Context
	tx  pgx.Tx
	row pgx.Row
}

func (r *tenantRow) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	if err != nil {
		_ = r.tx.Rollback(r.ctx)
		return err
	}
	if err := r.tx.Commit(r.ctx); err != nil {
		return err
	}
	return nil
}

type errRow struct {
	err error
}

func (r errRow) Scan(dest ...any) error {
	return r.err
}

type tenantRows struct {
	ctx       context.Context
	tx        pgx.Tx
	rows      pgx.Rows
	closeErr  error
	closeDone bool
}

func (r *tenantRows) Close() {
	if r.closeDone {
		return
	}
	r.closeDone = true
	r.rows.Close()
	if err := r.rows.Err(); err != nil {
		r.closeErr = err
		_ = r.tx.Rollback(r.ctx)
		return
	}
	r.closeErr = r.tx.Commit(r.ctx)
}

func (r *tenantRows) Err() error {
	if err := r.rows.Err(); err != nil {
		return err
	}
	return r.closeErr
}

func (r *tenantRows) CommandTag() pgconn.CommandTag {
	return r.rows.CommandTag()
}

func (r *tenantRows) FieldDescriptions() []pgconn.FieldDescription {
	return r.rows.FieldDescriptions()
}

func (r *tenantRows) Next() bool {
	return r.rows.Next()
}

func (r *tenantRows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *tenantRows) Values() ([]any, error) {
	return r.rows.Values()
}

func (r *tenantRows) RawValues() [][]byte {
	return r.rows.RawValues()
}

func (r *tenantRows) Conn() *pgx.Conn {
	return r.rows.Conn()
}
