package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stayflow/stayflow-track/internal/modules/dashboard/domain"
	"github.com/stayflow/stayflow-track/internal/platform/database"
)

type Repository struct {
	pool *database.TenantPool
}

func New(pool *database.TenantPool) *Repository {
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
			COALESCE(SUM(CASE WHEN p.created_at >= $3 THEN p.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN p.created_at >= $4 THEN p.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN p.created_at >= $5 THEN p.amount ELSE 0 END), 0)
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

// GetRevenueTrendByRange returns revenue data between specific dates.
func (r *Repository) GetRevenueTrendByRange(ctx context.Context, tenantID, propertyID uuid.UUID, startDate, endDate string) ([]domain.RevenueTrend, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT d::DATE, COALESCE(SUM(p.amount), 0)
		 FROM generate_series($3::DATE, $4::DATE, '1 day') AS d
		 LEFT JOIN payments p ON p.created_at::DATE = d::DATE
		   AND p.folio_id IN (SELECT id FROM folios WHERE tenant_id = $1 AND property_id = $2)
		   AND p.payment_type IN ('payment', 'deposit')
		 GROUP BY d::DATE ORDER BY d::DATE`,
		tenantID, propertyID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("revenue trend by range: %w", err)
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

// GetDailyCollection returns payment totals grouped by method for a given date.
func (r *Repository) GetDailyCollection(ctx context.Context, tenantID, propertyID uuid.UUID, date string) (*domain.DailyCollection, error) {
	var targetDate time.Time
	if date == "today" || date == "" {
		targetDate = time.Now().Truncate(24 * time.Hour)
	} else {
		var err error
		targetDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			targetDate = time.Now().Truncate(24 * time.Hour)
		}
	}
	nextDay := targetDate.AddDate(0, 0, 1)

	dc := &domain.DailyCollection{Date: targetDate.Format("2006-01-02")}

	err := r.pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN payment_type != 'refund' THEN amount ELSE 0 END), 0) AS total_collected,
			COALESCE(SUM(CASE WHEN payment_method = 'cash' AND payment_type != 'refund' THEN amount ELSE 0 END), 0) AS cash,
			COALESCE(SUM(CASE WHEN payment_method = 'upi' AND payment_type != 'refund' THEN amount ELSE 0 END), 0) AS upi,
			COALESCE(SUM(CASE WHEN payment_method = 'card' AND payment_type != 'refund' THEN amount ELSE 0 END), 0) AS card,
			COALESCE(SUM(CASE WHEN payment_method = 'bank_transfer' AND payment_type != 'refund' THEN amount ELSE 0 END), 0) AS bank_transfer,
			COALESCE(SUM(CASE WHEN payment_method = 'cheque' AND payment_type != 'refund' THEN amount ELSE 0 END), 0) AS cheque,
			COALESCE(SUM(CASE WHEN payment_type = 'refund' THEN amount ELSE 0 END), 0) AS refunds,
			COUNT(*) AS transactions
		 FROM payments p
		 JOIN folios f ON p.folio_id = f.id
		 WHERE f.tenant_id = $1 AND f.property_id = $2
		   AND p.created_at >= $3 AND p.created_at < $4`,
		tenantID, propertyID, targetDate, nextDay,
	).Scan(&dc.TotalCollected, &dc.Cash, &dc.UPI, &dc.Card, &dc.BankTransfer, &dc.Cheque, &dc.TotalRefunds, &dc.Transactions)
	if err != nil {
		return nil, fmt.Errorf("daily collection: %w", err)
	}

	dc.NetCollection = dc.TotalCollected.Sub(dc.TotalRefunds)
	return dc, nil
}

// GetOutstandingDues returns all folios/reservations with balance > 0.
func (r *Repository) GetOutstandingDues(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.OutstandingReport, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT
			r.id, g.first_name || ' ' || g.last_name, u.unit_number,
			r.check_in_date, r.check_out_date,
			f.total_amount, f.paid_amount, f.balance, r.status
		 FROM folios f
		 JOIN reservations r ON f.reservation_id = r.id
		 JOIN guests g ON r.guest_id = g.id
		 JOIN units u ON r.unit_id = u.id
		 WHERE f.tenant_id = $1 AND f.property_id = $2 AND f.balance > 0
		 ORDER BY f.balance DESC`,
		tenantID, propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("outstanding dues: %w", err)
	}
	defer rows.Close()

	report := &domain.OutstandingReport{}
	for rows.Next() {
		var d domain.OutstandingDue
		if err := rows.Scan(&d.ReservationID, &d.GuestName, &d.UnitNumber,
			&d.CheckInDate, &d.CheckOutDate, &d.TotalAmount, &d.PaidAmount, &d.Balance, &d.Status); err != nil {
			return nil, err
		}
		report.Dues = append(report.Dues, d)
		report.TotalOutstanding = report.TotalOutstanding.Add(d.Balance)
		report.Count++
	}
	if report.Dues == nil {
		report.Dues = []domain.OutstandingDue{}
	}
	return report, nil
}

// GetEndOfDaySummary returns combined end-of-day data for night audit.
func (r *Repository) GetEndOfDaySummary(ctx context.Context, tenantID, propertyID uuid.UUID, date string) (*domain.EndOfDaySummary, error) {
	summary := &domain.EndOfDaySummary{Date: date}

	// Occupancy
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'occupied')
		 FROM units WHERE tenant_id = $1 AND property_id = $2`,
		tenantID, propertyID,
	).Scan(&summary.TotalUnits, &summary.OccupiedUnits)
	if err != nil {
		return nil, fmt.Errorf("eod occupancy: %w", err)
	}
	if summary.TotalUnits > 0 {
		summary.OccupancyRate = float64(summary.OccupiedUnits) / float64(summary.TotalUnits) * 100
	}

	// Operations counts for the date
	err = r.pool.QueryRow(ctx,
		`SELECT
			COUNT(*) FILTER (WHERE actual_check_in::date = $3::date) AS check_ins,
			COUNT(*) FILTER (WHERE actual_check_out::date = $3::date) AS check_outs,
			COUNT(*) FILTER (WHERE booking_source = 'walk_in' AND actual_check_in::date = $3::date) AS walk_ins,
			COUNT(*) FILTER (WHERE is_no_show = true AND no_show_at::date = $3::date) AS no_shows,
			COUNT(*) FILTER (WHERE status = 'cancelled' AND cancelled_at::date = $3::date) AS cancellations
		 FROM reservations WHERE tenant_id = $1 AND property_id = $2`,
		tenantID, propertyID, date,
	).Scan(&summary.CheckIns, &summary.CheckOuts, &summary.WalkIns, &summary.NoShows, &summary.Cancellations)
	if err != nil {
		return nil, fmt.Errorf("eod operations: %w", err)
	}

	// Extensions count
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM stay_extensions WHERE tenant_id = $1 AND created_at::date = $2::date`,
		tenantID, date,
	).Scan(&summary.Extensions)

	// Collection
	collection, err := r.GetDailyCollection(ctx, tenantID, propertyID, date)
	if err != nil {
		return nil, err
	}
	summary.Collection = *collection

	// Outstanding
	outstanding, err := r.GetOutstandingDues(ctx, tenantID, propertyID)
	if err != nil {
		return nil, err
	}
	summary.Outstanding = *outstanding

	// Check if day is closed
	var closed bool
	var closedBy *uuid.UUID
	var closedAt *time.Time
	err = r.pool.QueryRow(ctx,
		`SELECT status = 'closed', closed_by, closed_at FROM night_audits
		 WHERE tenant_id = $1 AND property_id = $2 AND audit_date = $3::date`,
		tenantID, propertyID, date,
	).Scan(&closed, &closedBy, &closedAt)
	if err == nil {
		summary.IsClosed = closed
		summary.ClosedBy = closedBy
		summary.ClosedAt = closedAt
	}

	return summary, nil
}

// CloseDay marks the day as audited/closed.
func (r *Repository) CloseDay(ctx context.Context, tenantID, propertyID, userID uuid.UUID, date string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO night_audits (tenant_id, property_id, audit_date, status, closed_by, closed_at,
			total_units, occupied_units, occupancy_rate, check_ins, check_outs)
		 VALUES ($1, $2, $3::date, 'closed', $4, NOW(), 0, 0, 0, 0, 0)
		 ON CONFLICT (tenant_id, property_id, audit_date) DO UPDATE SET
			status = 'closed', closed_by = $4, closed_at = NOW()`,
		tenantID, propertyID, date, userID,
	)
	return err
}
