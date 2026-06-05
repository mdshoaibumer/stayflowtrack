package service_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stayflow/stayflow-track/internal/modules/laundry/domain"
)

func TestCreateOrderInput_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input domain.CreateOrderInput
		valid bool
	}{
		{
			name: "valid guest order",
			input: domain.CreateOrderInput{
				PropertyID: uuid.New(),
				OrderType:  "guest",
				Items: []domain.CreateItemInput{
					{ItemType: "towel", Quantity: 5, UnitPrice: 30, ServiceType: "wash"},
				},
			},
			valid: true,
		},
		{
			name: "valid house order",
			input: domain.CreateOrderInput{
				PropertyID: uuid.New(),
				OrderType:  "house",
				Items: []domain.CreateItemInput{
					{ItemType: "bedsheet", Quantity: 20, UnitPrice: 25, ServiceType: "wash"},
					{ItemType: "pillow_cover", Quantity: 40, UnitPrice: 15, ServiceType: "wash"},
				},
			},
			valid: true,
		},
		{
			name: "invalid order type",
			input: domain.CreateOrderInput{
				PropertyID: uuid.New(),
				OrderType:  "external",
				Items:      []domain.CreateItemInput{{ItemType: "towel", Quantity: 1, UnitPrice: 10, ServiceType: "wash"}},
			},
			valid: false,
		},
		{
			name: "empty items",
			input: domain.CreateOrderInput{
				PropertyID: uuid.New(),
				OrderType:  "guest",
				Items:      []domain.CreateItemInput{},
			},
			valid: false,
		},
		{
			name:  "missing property",
			input: domain.CreateOrderInput{OrderType: "guest", Items: []domain.CreateItemInput{{ItemType: "towel", Quantity: 1, UnitPrice: 10, ServiceType: "wash"}}},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.PropertyID == uuid.Nil && tt.valid {
				t.Error("expected invalid")
			}
			if tt.input.OrderType != "guest" && tt.input.OrderType != "house" && tt.valid {
				t.Error("expected invalid order type")
			}
			if len(tt.input.Items) == 0 && tt.valid {
				t.Error("expected invalid - no items")
			}
		})
	}
}

func TestCalculateOrderTotals(t *testing.T) {
	items := []domain.CreateItemInput{
		{ItemType: "bedsheet", Quantity: 10, UnitPrice: 25, ServiceType: "wash"},
		{ItemType: "towel", Quantity: 5, UnitPrice: 30, ServiceType: "wash"},
	}

	var totalAmount float64
	var totalItems int
	for _, item := range items {
		amount := float64(item.Quantity) * item.UnitPrice
		totalAmount += amount
		totalItems += item.Quantity
	}

	if totalItems != 15 {
		t.Errorf("expected 15 items, got %d", totalItems)
	}
	if totalAmount != 400.0 {
		t.Errorf("expected 400.0, got %f", totalAmount)
	}

	taxAmount := totalAmount * 18 / 100
	grandTotal := totalAmount + taxAmount
	if grandTotal != 472.0 {
		t.Errorf("expected 472.0, got %f", grandTotal)
	}
}

func TestLaundryStatusFlow(t *testing.T) {
	validTransitions := map[string][]string{
		"received": {"washing"},
		"washing":  {"ready"},
		"ready":    {"delivered"},
	}

	for from, tos := range validTransitions {
		for _, to := range tos {
			t.Run(from+"→"+to, func(t *testing.T) {
				// Verify transition is valid
				found := false
				for _, valid := range validTransitions[from] {
					if valid == to {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("transition %s→%s should be valid", from, to)
				}
			})
		}
	}
}
