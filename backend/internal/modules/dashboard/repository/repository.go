package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stayflow/stayflow-track/internal/modules/dashboard/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetOccupancy(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.OccupancyMetric, error) {
	var m domain.OccupancyMetric
	err := r.pool.QueryRow(ctx,
		`SELECT
			COUNT(*) AS total_units,
			COUNT(*) FILTER (WHERE status = 'occupied') AS occupied,
			COUNT(*) FILTER (WHERE status = 'available') AS available
		 FROM units WHERE tenant_id = $1 AND property_id = $2`,
		tenantID, propertyID,
	).Scan(&m.TotalUnits, &m.OccupiedUnits, &m.AvailableUnits)
	if err != nil {
		return nil, fmt.Errorf("get occupancy: %w", err)
	}
	if m.TotalUnits > 0 {
		m.OccupancyRate = float64(m.OccupiedUnits) / float64(m.TotalUnits) * 100
	}
	return &m, nil
}

func (r *Repository) GetRevenue(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.RevenueMetric, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := startOfDay.AddDate(0, 0, -int(now.Weekday()))
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var m domain.RevenueMetric
	m.Currency = "INR"

	err := r.pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN created_at >= $3 THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN created_at >= $4 THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN created_at >= $5 THEN amount ELSE 0 END), 0)
		 FROM payments p
		 JOIN folios f ON p.folio_id = f.id
		 WHERE f.tenant_id = $1 AND f.property_id = $2
		   AND p.payment_type IN ('payment', 'deposit')`,
		tenantID, propertyID, startOfDay, startOfWeek, startOfMonth,
	).Scan(&m.Today, &m.ThisWeek, &m.ThisMonth)
	if err != nil {
		return nil, fmt.Errorf("get revenue: %w", err)
	}
	return &m, nil
}

func (r *Repository) GetOperations(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.OperationsMetric, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)

	var m domain.OperationsMetric
	err := r.pool.QueryRow(ctx,
		`SELECT
			COUNT(*) FILTER (WHERE status = 'checked_in' AND actual_check_in >= $3 AND actual_check_in < $4),
			COUNT(*) FILTER (WHERE status = 'checked_out' AND actual_check_out >= $3 AND actual_check_out < $4),
			COUNT(*) FILTER (WHERE status = 'confirmed' AND check_in_date = $3::DATE),
			COUNT(*) FILTER (WHERE status = 'checked_in' AND check_out_date = $3::DATE)
		 FROM reservations WHERE tenant_id = $1 AND property_id = $2`,
		tenantID, propertyID, today, tomorrow,
	).Scan(&m.CheckInsToday, &m.CheckOutsToday, &m.ExpectedArrivals, &m.ExpectedDeparts)
	if err != nil {
		return nil, fmt.Errorf("get operations: %w", err)
	}
	return &m, nil
}

func (r *Repository) GetPendingPayments(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.PaymentMetric, error) {
	var m domain.PaymentMetric
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(balance), 0)
		 FROM folios WHERE tenant_id = $1 AND property_id = $2 AND status = 'open' AND balance > 0`,
		tenantID, propertyID,
	).Scan(&m.PendingCount, &m.PendingAmount)
	if err != nil {
		return nil, fmt.Errorf("get pending payments: %w", err)
	}

	// Overdue invoices
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM invoices
		 WHERE tenant_id = $1 AND property_id = $2 AND status IN ('issued', 'partially_paid') AND due_date < NOW()`,
		tenantID, propertyID,
	).Scan(&m.OverdueCount)

	return &m, nil
}

func (r *Repository) GetHousekeepingStats(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.StatusCounts, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT status::TEXT, COUNT(*)::INT FROM housekeeping_tasks
		 WHERE tenant_id = $1 AND property_id = $2 AND completed_at IS NULL
		 GROUP BY status`,
		tenantID, propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("housekeeping stats: %w", err)
	}
	defer rows.Close()

	sc := &domain.StatusCounts{Counts: map[string]int{}}
	for rows.Next() {
		var s string
		var c int
		if err := rows.Scan(&s, &c); err != nil {
			return nil, err
		}
		sc.Counts[s] = c
		sc.Total += c
	}
	return sc, nil
}

func (r *Repository) GetLaundryStats(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.StatusCounts, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT status::TEXT, COUNT(*)::INT FROM laundry_orders
		 WHERE tenant_id = $1 AND property_id = $2 AND status != 'delivered'
		 GROUP BY status`,
		tenantID, propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("laundry stats: %w", err)
	}
	defer rows.Close()

	sc := &domain.StatusCounts{Counts: map[string]int{}}
	for rows.Next() {
		var s string
		var c int
		if err := rows.Scan(&s, &c); err != nil {
			return nil, err
		}
		sc.Counts[s] = c
		sc.Total += c
	}
	return sc, nil
}

func (r *Repository) GetRevenueTrend(ctx context.Context, tenantID, propertyID uuid.UUID, days int) ([]domain.RevenueTrend, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT d::DATE, COALESCE(SUM(p.amount), 0)
		 FROM generate_series(NOW() - ($3::INT || ' days')::INTERVAL, NOW(), '1 day') AS d
		 LEFT JOIN payments p ON p.created_at::DATE = d::DATE
		   AND p.folio_id IN (SELECT id FROM folios WHERE tenant_id = $1 AND property_id = $2)
		   AND p.payment_type IN ('payment', 'deposit')
		 GROUP BY d::DATE ORDER BY d::DATE`,
		tenantID, propertyID, days,
	)
	if err != nil {
		return nil, fmt.Errorf("revenue trend: %w", err)
	}
	defer rows.Close()

	var trends []domain.RevenueTrend
	for rows.Next() {
		var t domain.RevenueTrend
		if err := rows.Scan(&t.Date, &t.Revenue); err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}
	if trends == nil {
		trends = []domain.RevenueTrend{}
	}
	return trends, nil
}
