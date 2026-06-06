package domain

import (
	"time"

	"github.com/google/uuid"
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

// DailyCollection shows payment breakdown by method for a given day.
type DailyCollection struct {
	Date           string          `json:"date"`
	TotalCollected decimal.Decimal `json:"total_collected"`
	Cash           decimal.Decimal `json:"cash"`
	UPI            decimal.Decimal `json:"upi"`
	Card           decimal.Decimal `json:"card"`
	BankTransfer   decimal.Decimal `json:"bank_transfer"`
	Cheque         decimal.Decimal `json:"cheque"`
	TotalRefunds   decimal.Decimal `json:"total_refunds"`
	NetCollection  decimal.Decimal `json:"net_collection"`
	Transactions   int             `json:"transactions"`
}

// OutstandingDue represents a guest with pending balance.
type OutstandingDue struct {
	ReservationID uuid.UUID       `json:"reservation_id"`
	GuestName     string          `json:"guest_name"`
	UnitNumber    string          `json:"unit_number"`
	CheckInDate   time.Time       `json:"check_in_date"`
	CheckOutDate  time.Time       `json:"check_out_date"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	PaidAmount    decimal.Decimal `json:"paid_amount"`
	Balance       decimal.Decimal `json:"balance"`
	Status        string          `json:"status"`
}

// OutstandingReport aggregates all pending dues.
type OutstandingReport struct {
	TotalOutstanding decimal.Decimal  `json:"total_outstanding"`
	Count            int              `json:"count"`
	Dues             []OutstandingDue `json:"dues"`
}

// EndOfDaySummary is the comprehensive night audit report.
type EndOfDaySummary struct {
	Date string `json:"date"`
	// Occupancy
	TotalUnits    int     `json:"total_units"`
	OccupiedUnits int     `json:"occupied_units"`
	OccupancyRate float64 `json:"occupancy_rate"`
	// Operations
	CheckIns      int `json:"check_ins"`
	CheckOuts     int `json:"check_outs"`
	WalkIns       int `json:"walk_ins"`
	NoShows       int `json:"no_shows"`
	Cancellations int `json:"cancellations"`
	Extensions    int `json:"extensions"`
	// Collection
	Collection DailyCollection `json:"collection"`
	// Outstanding
	Outstanding OutstandingReport `json:"outstanding"`
	// Status
	IsClosed bool       `json:"is_closed"`
	ClosedBy *uuid.UUID `json:"closed_by,omitempty"`
	ClosedAt *time.Time `json:"closed_at,omitempty"`
}
