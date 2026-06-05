package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// DashboardMetrics contains all top-level KPIs.
type DashboardMetrics struct {
	Date            string           `json:"date"`
	Occupancy       OccupancyMetric  `json:"occupancy"`
	Revenue         RevenueMetric    `json:"revenue"`
	Operations      OperationsMetric `json:"operations"`
	Housekeeping    StatusCounts     `json:"housekeeping"`
	Laundry         StatusCounts     `json:"laundry"`
	PendingPayments PaymentMetric    `json:"pending_payments"`
}

type OccupancyMetric struct {
	TotalUnits     int     `json:"total_units"`
	OccupiedUnits  int     `json:"occupied_units"`
	AvailableUnits int     `json:"available_units"`
	OccupancyRate  float64 `json:"occupancy_rate"`
}

type RevenueMetric struct {
	Today     decimal.Decimal `json:"today"`
	ThisWeek  decimal.Decimal `json:"this_week"`
	ThisMonth decimal.Decimal `json:"this_month"`
	Currency  string          `json:"currency"`
}

type OperationsMetric struct {
	CheckInsToday    int `json:"check_ins_today"`
	CheckOutsToday   int `json:"check_outs_today"`
	ExpectedArrivals int `json:"expected_arrivals"`
	ExpectedDeparts  int `json:"expected_departures"`
}

type StatusCounts struct {
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
}

type PaymentMetric struct {
	PendingCount  int             `json:"pending_count"`
	PendingAmount decimal.Decimal `json:"pending_amount"`
	OverdueCount  int             `json:"overdue_count"`
}

// RevenueTrend for chart data.
type RevenueTrend struct {
	Date    time.Time       `json:"date"`
	Revenue decimal.Decimal `json:"revenue"`
}

// OccupancyTrend for chart data.
type OccupancyTrend struct {
	Date          time.Time `json:"date"`
	OccupancyRate float64   `json:"occupancy_rate"`
}
