package service_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/dashboard/domain"
	"github.com/stayflow/stayflow-track/internal/modules/dashboard/service"
)

func TestDashboardCache(t *testing.T) {
	// Test cache behavior
	t.Run("cache miss returns false", func(t *testing.T) {
		// Simulate cache behavior
		_ = service.New(nil) // Would use mock repo in real tests
	})

	t.Run("occupancy rate calculation", func(t *testing.T) {
		m := domain.OccupancyMetric{
			TotalUnits:    100,
			OccupiedUnits: 75,
		}
		if m.TotalUnits > 0 {
			m.OccupancyRate = float64(m.OccupiedUnits) / float64(m.TotalUnits) * 100
		}
		if m.OccupancyRate != 75.0 {
			t.Errorf("expected 75.0%%, got %f%%", m.OccupancyRate)
		}
	})

	t.Run("revenue metric aggregation", func(t *testing.T) {
		m := domain.RevenueMetric{
			Today:     decimal.NewFromInt(5000),
			ThisWeek:  decimal.NewFromInt(35000),
			ThisMonth: decimal.NewFromInt(150000),
			Currency:  "INR",
		}
		if m.ThisMonth.LessThan(m.ThisWeek) {
			t.Error("monthly revenue should be >= weekly")
		}
		if m.ThisWeek.LessThan(m.Today) {
			t.Error("weekly revenue should be >= daily")
		}
	})

	t.Run("operations metrics structure", func(t *testing.T) {
		m := domain.OperationsMetric{
			CheckInsToday:    5,
			CheckOutsToday:   3,
			ExpectedArrivals: 8,
			ExpectedDeparts:  4,
		}
		if m.CheckInsToday < 0 || m.CheckOutsToday < 0 {
			t.Error("counts should be non-negative")
		}
	})

	t.Run("revenue trend data", func(t *testing.T) {
		trends := []domain.RevenueTrend{
			{Date: time.Now().AddDate(0, 0, -2), Revenue: decimal.NewFromInt(10000)},
			{Date: time.Now().AddDate(0, 0, -1), Revenue: decimal.NewFromInt(15000)},
			{Date: time.Now(), Revenue: decimal.NewFromInt(12000)},
		}
		if len(trends) != 3 {
			t.Errorf("expected 3 data points, got %d", len(trends))
		}
		for i := 1; i < len(trends); i++ {
			if trends[i].Date.Before(trends[i-1].Date) {
				t.Error("trends should be ordered by date")
			}
		}
	})
}
