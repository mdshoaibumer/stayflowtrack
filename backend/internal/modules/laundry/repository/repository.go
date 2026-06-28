package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/laundry/domain"
	"github.com/stayflow/stayflow-track/internal/platform/database"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *database.TenantPool
}

func New(pool *database.TenantPool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Pool() *database.TenantPool {
	return r.pool
}

func (r *Repository) CreateOrder(ctx context.Context, order *domain.LaundryOrder, items []domain.LaundryItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx,
		`INSERT INTO laundry_orders (tenant_id, property_id, reservation_id, guest_id, folio_id, unit_id,
		  order_number, order_type, status, total_items, total_amount, tax_amount, grand_total, notes, received_by)
		 VALUES ($1, $2, $3, $4, $5, $6, generate_laundry_order_number($1), $7, 'received', $8, $9, $10, $11, $12, $13)
		 RETURNING id, order_number, received_at, created_at, updated_at`,
		order.TenantID, order.PropertyID, order.ReservationID, order.GuestID, order.FolioID, order.UnitID,
		order.OrderType, order.TotalItems, order.TotalAmount, order.TaxAmount, order.GrandTotal,
		order.Notes, order.ReceivedBy,
	).Scan(&order.ID, &order.OrderNumber, &order.ReceivedAt, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for i := range items {
		items[i].OrderID = order.ID
		items[i].Amount = decimal.NewFromInt(int64(items[i].Quantity)).Mul(items[i].UnitPrice)
		err = tx.QueryRow(ctx,
			`INSERT INTO laundry_items (order_id, item_type, description, quantity, unit_price, amount, service_type)
			 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`,
			items[i].OrderID, items[i].ItemType, items[i].Description,
			items[i].Quantity, items[i].UnitPrice, items[i].Amount, items[i].ServiceType,
		).Scan(&items[i].ID, &items[i].CreatedAt)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}
	order.Items = items

	return tx.Commit(ctx)
}

func (r *Repository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.LaundryOrder, error) {
	var o domain.LaundryOrder
	err := r.pool.QueryRow(ctx,
		`SELECT lo.id, lo.tenant_id, lo.property_id, lo.reservation_id, lo.guest_id, lo.folio_id, lo.unit_id,
		        lo.order_number, lo.order_type, lo.status, lo.total_items, lo.total_amount, lo.tax_amount, lo.grand_total,
		        COALESCE(lo.notes, ''), lo.received_by, lo.received_at, lo.washed_at, lo.ready_at, lo.delivered_at, lo.delivered_by,
		        lo.posted_to_folio, lo.created_at, lo.updated_at,
		        COALESCE(g.first_name || ' ' || g.last_name, ''),
		        COALESCE(u.unit_number, '')
		 FROM laundry_orders lo
		 LEFT JOIN guests g ON lo.guest_id = g.id
		 LEFT JOIN units u ON lo.unit_id = u.id
		 WHERE lo.id = $1 AND lo.tenant_id = $2`,
		id, tenantID,
	).Scan(&o.ID, &o.TenantID, &o.PropertyID, &o.ReservationID, &o.GuestID, &o.FolioID, &o.UnitID,
		&o.OrderNumber, &o.OrderType, &o.Status, &o.TotalItems, &o.TotalAmount, &o.TaxAmount, &o.GrandTotal,
		&o.Notes, &o.ReceivedBy, &o.ReceivedAt, &o.WashedAt, &o.ReadyAt, &o.DeliveredAt, &o.DeliveredBy,
		&o.PostedToFolio, &o.CreatedAt, &o.UpdatedAt,
		&o.GuestName, &o.UnitNumber)
	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("laundry_order", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	// Fetch items
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, item_type, COALESCE(description, ''), quantity, unit_price, amount, service_type, created_at
		 FROM laundry_items WHERE order_id = $1 ORDER BY created_at`, id)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.LaundryItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ItemType, &item.Description,
			&item.Quantity, &item.UnitPrice, &item.Amount, &item.ServiceType, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		o.Items = append(o.Items, item)
	}
	if o.Items == nil {
		o.Items = []domain.LaundryItem{}
	}

	return &o, nil
}

func (r *Repository) ListOrders(ctx context.Context, tenantID uuid.UUID, propertyID *uuid.UUID, status string, limit, offset int) ([]domain.LaundryOrder, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM laundry_orders
		 WHERE tenant_id = $1
		   AND ($2::UUID IS NULL OR property_id = $2::UUID)
		   AND ($3::VARCHAR = '' OR status::TEXT = $3)`,
		tenantID, propertyID, status,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT lo.id, lo.tenant_id, lo.property_id, lo.reservation_id, lo.guest_id, lo.folio_id, lo.unit_id,
		        lo.order_number, lo.order_type, lo.status, lo.total_items, lo.total_amount, lo.tax_amount, lo.grand_total,
		        COALESCE(lo.notes, ''), lo.received_by, lo.received_at, lo.washed_at, lo.ready_at, lo.delivered_at, lo.delivered_by,
		        lo.posted_to_folio, lo.created_at, lo.updated_at,
		        COALESCE(g.first_name || ' ' || g.last_name, ''),
		        COALESCE(u.unit_number, '')
		 FROM laundry_orders lo
		 LEFT JOIN guests g ON lo.guest_id = g.id
		 LEFT JOIN units u ON lo.unit_id = u.id
		 WHERE lo.tenant_id = $1
		   AND ($2::UUID IS NULL OR lo.property_id = $2::UUID)
		   AND ($3::VARCHAR = '' OR lo.status::TEXT = $3)
		 ORDER BY lo.created_at DESC LIMIT $4 OFFSET $5`,
		tenantID, propertyID, status, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.LaundryOrder
	for rows.Next() {
		var o domain.LaundryOrder
		if err := rows.Scan(&o.ID, &o.TenantID, &o.PropertyID, &o.ReservationID, &o.GuestID, &o.FolioID, &o.UnitID,
			&o.OrderNumber, &o.OrderType, &o.Status, &o.TotalItems, &o.TotalAmount, &o.TaxAmount, &o.GrandTotal,
			&o.Notes, &o.ReceivedBy, &o.ReceivedAt, &o.WashedAt, &o.ReadyAt, &o.DeliveredAt, &o.DeliveredBy,
			&o.PostedToFolio, &o.CreatedAt, &o.UpdatedAt,
			&o.GuestName, &o.UnitNumber); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []domain.LaundryOrder{}
	}
	return orders, count, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, orderID, tenantID, userID uuid.UUID, status string) error {
	now := time.Now()
	var err error

	switch status {
	case "washing":
		_, err = r.pool.Exec(ctx,
			`UPDATE laundry_orders SET status = $3, washed_at = $4 WHERE id = $1 AND tenant_id = $2`,
			orderID, tenantID, status, now)
	case "ready":
		_, err = r.pool.Exec(ctx,
			`UPDATE laundry_orders SET status = $3, ready_at = $4 WHERE id = $1 AND tenant_id = $2`,
			orderID, tenantID, status, now)
	case "delivered":
		_, err = r.pool.Exec(ctx,
			`UPDATE laundry_orders SET status = $3, delivered_at = $4, delivered_by = $5 WHERE id = $1 AND tenant_id = $2`,
			orderID, tenantID, status, now, userID)
	default:
		_, err = r.pool.Exec(ctx,
			`UPDATE laundry_orders SET status = $3 WHERE id = $1 AND tenant_id = $2`,
			orderID, tenantID, status)
	}
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// PostToFolio posts laundry charges to the guest folio.
func (r *Repository) PostToFolio(ctx context.Context, orderID, tenantID, userID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var folioID uuid.UUID
	var grandTotal, taxAmount float64
	var orderNumber string
	var posted bool

	err = tx.QueryRow(ctx,
		`SELECT folio_id, grand_total, tax_amount, order_number, posted_to_folio
		 FROM laundry_orders WHERE id = $1 AND tenant_id = $2`,
		orderID, tenantID,
	).Scan(&folioID, &grandTotal, &taxAmount, &orderNumber, &posted)
	if err == pgx.ErrNoRows {
		return apperrors.NotFound("laundry_order", orderID.String())
	}
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}
	if posted {
		return apperrors.BadRequest("charges already posted to folio")
	}
	if folioID == uuid.Nil {
		return apperrors.BadRequest("no folio linked to this laundry order")
	}

	// Add line item to folio
	amount := grandTotal - taxAmount
	taxRate := 18.0
	if taxAmount > 0 && amount > 0 {
		taxRate = (taxAmount / amount) * 100
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO line_items (tenant_id, folio_id, category, description, quantity, unit_price, amount, tax_rate, tax_amount, total, date, created_by)
		 VALUES ($1, $2, 'laundry', $3, 1, $4, $4, $5, $6, $7, NOW(), $8)`,
		tenantID, folioID,
		fmt.Sprintf("Laundry - %s", orderNumber),
		amount, taxRate, taxAmount, grandTotal, userID,
	)
	if err != nil {
		return fmt.Errorf("insert line item: %w", err)
	}

	// Update folio totals
	_, err = tx.Exec(ctx,
		`UPDATE folios SET subtotal = subtotal + $2, tax_total = tax_total + $3, total_amount = total_amount + $4, balance = balance + $4 WHERE id = $1`,
		folioID, amount, taxAmount, grandTotal,
	)
	if err != nil {
		return fmt.Errorf("update folio: %w", err)
	}

	// Mark as posted
	_, err = tx.Exec(ctx,
		`UPDATE laundry_orders SET posted_to_folio = true WHERE id = $1`, orderID)
	if err != nil {
		return fmt.Errorf("mark posted: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetStats(ctx context.Context, tenantID, propertyID uuid.UUID) (map[string]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT status::TEXT, COUNT(*)::INT FROM laundry_orders
		 WHERE tenant_id = $1 AND property_id = $2 AND status != 'delivered'
		 GROUP BY status`,
		tenantID, propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	defer rows.Close()

	stats := map[string]int{"received": 0, "washing": 0, "ready": 0}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, nil
}

// CreateRateCard saves a default item + price for quick laundry ordering.
func (r *Repository) CreateRateCard(ctx context.Context, tenantID uuid.UUID, input domain.CreateRateCardInput) (*domain.LaundryRateCard, error) {
	var card domain.LaundryRateCard
	err := r.pool.QueryRow(ctx,
		`INSERT INTO laundry_rate_cards (tenant_id, property_id, item_type, item_name, default_rate, service_type, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, true) RETURNING id, tenant_id, property_id, item_type, item_name, default_rate, service_type, is_active, created_at, updated_at`,
		tenantID, input.PropertyID, input.ItemType, input.ItemName, input.DefaultRate, input.ServiceType,
	).Scan(&card.ID, &card.TenantID, &card.PropertyID, &card.ItemType, &card.ItemName, &card.DefaultRate, &card.ServiceType, &card.IsActive, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create rate card: %w", err)
	}
	return &card, nil
}

// ListRateCards returns all rate card items for a property.
func (r *Repository) ListRateCards(ctx context.Context, tenantID, propertyID uuid.UUID) ([]domain.LaundryRateCard, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, property_id, item_type, item_name, default_rate, service_type, is_active, created_at, updated_at
		 FROM laundry_rate_cards WHERE tenant_id = $1 AND property_id = $2 AND is_active = true ORDER BY item_name`,
		tenantID, propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list rate cards: %w", err)
	}
	defer rows.Close()

	var cards []domain.LaundryRateCard
	for rows.Next() {
		var c domain.LaundryRateCard
		if err := rows.Scan(&c.ID, &c.TenantID, &c.PropertyID, &c.ItemType, &c.ItemName, &c.DefaultRate, &c.ServiceType, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	if cards == nil {
		cards = []domain.LaundryRateCard{}
	}
	return cards, nil
}

// UpdateRateCard modifies an existing rate card.
func (r *Repository) UpdateRateCard(ctx context.Context, tenantID uuid.UUID, input domain.UpdateRateCardInput) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE laundry_rate_cards SET
			item_name = COALESCE(NULLIF($3, ''), item_name),
			default_rate = CASE WHEN $4 > 0 THEN $4 ELSE default_rate END,
			is_active = COALESCE($5, is_active),
			updated_at = NOW()
		 WHERE id = $2 AND tenant_id = $1`,
		tenantID, input.ID, input.ItemName, input.DefaultRate, input.IsActive,
	)
	if err != nil {
		return fmt.Errorf("update rate card: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("rate card", input.ID.String())
	}
	return nil
}
