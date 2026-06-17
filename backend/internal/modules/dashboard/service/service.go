package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stayflow/stayflow-track/internal/modules/dashboard/domain"
	"github.com/stayflow/stayflow-track/internal/modules/dashboard/repository"
)

type Service struct {
	repo  *repository.Repository
	cache *dashboardCache
}

func New(repo *repository.Repository) *Service {
	return &Service{
		repo:  repo,
		cache: newCache(30 * time.Second),
	}
}

// GetDashboard returns all metrics with caching.
func (s *Service) GetDashboard(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.DashboardMetrics, error) {
	key := tenantID.String() + ":" + propertyID.String()

	if cached, ok := s.cache.get(key); ok {
		return cached, nil
	}

	var (
		wg                                                sync.WaitGroup
		occupancy                                         *domain.OccupancyMetric
		revenue                                           *domain.RevenueMetric
		ops                                               *domain.OperationsMetric
		hk                                                *domain.StatusCounts
		laundry                                           *domain.StatusCounts
		payments                                          *domain.PaymentMetric
		errOcc, errRev, errOps, errHk, errLaundry, errPay error
	)

	wg.Add(6)
	go func() { defer wg.Done(); occupancy, errOcc = s.repo.GetOccupancy(ctx, tenantID, propertyID) }()
	go func() { defer wg.Done(); revenue, errRev = s.repo.GetRevenue(ctx, tenantID, propertyID) }()
	go func() { defer wg.Done(); ops, errOps = s.repo.GetOperations(ctx, tenantID, propertyID) }()
	go func() { defer wg.Done(); hk, errHk = s.repo.GetHousekeepingStats(ctx, tenantID, propertyID) }()
	go func() { defer wg.Done(); laundry, errLaundry = s.repo.GetLaundryStats(ctx, tenantID, propertyID) }()
	go func() { defer wg.Done(); payments, errPay = s.repo.GetPendingPayments(ctx, tenantID, propertyID) }()
	wg.Wait()

	for _, err := range []error{errOcc, errRev, errOps, errHk, errLaundry, errPay} {
		if err != nil {
			return nil, err
		}
	}

	metrics := &domain.DashboardMetrics{
		Date:            time.Now().Format("2006-01-02"),
		Occupancy:       *occupancy,
		Revenue:         *revenue,
		Operations:      *ops,
		Housekeeping:    *hk,
		Laundry:         *laundry,
		PendingPayments: *payments,
	}

	s.cache.set(key, metrics)
	return metrics, nil
}

func (s *Service) GetRevenueTrend(ctx context.Context, tenantID, propertyID uuid.UUID, days int) ([]domain.RevenueTrend, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	return s.repo.GetRevenueTrend(ctx, tenantID, propertyID, days)
}

// GetRevenueTrendByRange returns revenue data for a specific date range.
func (s *Service) GetRevenueTrendByRange(ctx context.Context, tenantID, propertyID uuid.UUID, startDate, endDate string) ([]domain.RevenueTrend, error) {
	return s.repo.GetRevenueTrendByRange(ctx, tenantID, propertyID, startDate, endDate)
}

// GetDailyCollection returns payment breakdown by method for a date.
func (s *Service) GetDailyCollection(ctx context.Context, tenantID, propertyID uuid.UUID, date string) (*domain.DailyCollection, error) {
	return s.repo.GetDailyCollection(ctx, tenantID, propertyID, date)
}

// GetOutstandingDues returns all folios with balance > 0.
func (s *Service) GetOutstandingDues(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.OutstandingReport, error) {
	return s.repo.GetOutstandingDues(ctx, tenantID, propertyID)
}

// GetEndOfDaySummary returns a comprehensive end-of-day report combining all key metrics.
func (s *Service) GetEndOfDaySummary(ctx context.Context, tenantID, propertyID uuid.UUID, date string) (*domain.EndOfDaySummary, error) {
	return s.repo.GetEndOfDaySummary(ctx, tenantID, propertyID, date)
}

// CloseDay closes the day audit, preventing further changes to that day's record.
func (s *Service) CloseDay(ctx context.Context, tenantID, propertyID, userID uuid.UUID, date string) error {
	return s.repo.CloseDay(ctx, tenantID, propertyID, userID, date)
}

// In-memory cache with TTL.
type dashboardCache struct {
	mu    sync.RWMutex
	items map[string]*cacheEntry
	ttl   time.Duration
}

type cacheEntry struct {
	data      *domain.DashboardMetrics
	expiresAt time.Time
}

func newCache(ttl time.Duration) *dashboardCache {
	c := &dashboardCache{
		items: make(map[string]*cacheEntry),
		ttl:   ttl,
	}
	// Periodic cleanup of expired entries to prevent unbounded memory growth
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			c.evictExpired()
		}
	}()
	return c
}

func (c *dashboardCache) get(key string) (*domain.DashboardMetrics, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.items[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func (c *dashboardCache) set(key string, data *domain.DashboardMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *dashboardCache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, entry := range c.items {
		if now.After(entry.expiresAt) {
			delete(c.items, key)
		}
	}
}
