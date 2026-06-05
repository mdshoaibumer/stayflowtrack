package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/laundry/domain"
	"github.com/stayflow/stayflow-track/internal/modules/laundry/repository"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrder(ctx context.Context, tenantID, userID uuid.UUID, input domain.CreateOrderInput) (*domain.LaundryOrder, error) {
	var totalItems int
	var totalAmount decimal.Decimal
	items := make([]domain.LaundryItem, 0, len(input.Items))

	for _, ii := range input.Items {
		unitPrice := decimal.NewFromFloat(ii.UnitPrice)
		amount := decimal.NewFromInt(int64(ii.Quantity)).Mul(unitPrice)
		totalItems += ii.Quantity
		totalAmount = totalAmount.Add(amount)
		items = append(items, domain.LaundryItem{
			ItemType:    ii.ItemType,
			Description: ii.Description,
			Quantity:    ii.Quantity,
			UnitPrice:   unitPrice,
			Amount:      amount,
			ServiceType: ii.ServiceType,
		})
	}

	taxRate := decimal.NewFromInt(18)
	taxAmount := totalAmount.Mul(taxRate).Div(decimal.NewFromInt(100)).Round(2)
	grandTotal := totalAmount.Add(taxAmount)

	// Resolve folio if reservation is provided
	var folioID *uuid.UUID
	if input.ReservationID != nil {
		var fid uuid.UUID
		// Try to find open folio for this reservation
		err := s.repo.Pool().QueryRow(ctx,
			`SELECT id FROM folios WHERE reservation_id = $1 AND tenant_id = $2 AND status = 'open' LIMIT 1`,
			input.ReservationID, tenantID).Scan(&fid)
		if err == nil {
			folioID = &fid
		}
	}

	order := &domain.LaundryOrder{
		TenantID:      tenantID,
		PropertyID:    input.PropertyID,
		ReservationID: input.ReservationID,
		GuestID:       input.GuestID,
		FolioID:       folioID,
		UnitID:        input.UnitID,
		OrderType:     input.OrderType,
		Status:        domain.LaundryReceived,
		TotalItems:    totalItems,
		TotalAmount:   totalAmount,
		TaxAmount:     taxAmount,
		GrandTotal:    grandTotal,
		Notes:         input.Notes,
		ReceivedBy:    userID,
	}

	if err := s.repo.CreateOrder(ctx, order, items); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *Service) GetOrder(ctx context.Context, orderID, tenantID uuid.UUID) (*domain.LaundryOrder, error) {
	return s.repo.GetByID(ctx, orderID, tenantID)
}

func (s *Service) ListOrders(ctx context.Context, tenantID uuid.UUID, propertyID *uuid.UUID, status string, limit, offset int) ([]domain.LaundryOrder, int64, error) {
	return s.repo.ListOrders(ctx, tenantID, propertyID, status, limit, offset)
}

func (s *Service) UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, input domain.UpdateStatusInput) error {
	return s.repo.UpdateStatus(ctx, input.OrderID, tenantID, userID, input.Status)
}

func (s *Service) PostToFolio(ctx context.Context, orderID, tenantID, userID uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, orderID, tenantID)
	if err != nil {
		return err
	}
	if order.OrderType == "house" {
		return apperrors.BadRequest("house laundry orders cannot be posted to guest folio")
	}
	return s.repo.PostToFolio(ctx, orderID, tenantID, userID)
}

func (s *Service) GetStats(ctx context.Context, tenantID, propertyID uuid.UUID) (map[string]int, error) {
	return s.repo.GetStats(ctx, tenantID, propertyID)
}
