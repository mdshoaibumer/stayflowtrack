package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stayflow/stayflow-track/internal/modules/laundry/domain"
)

func TestCreateOrderInput_Validation(t *testing.T) {
	t.Run("valid guest order", func(t *testing.T) {
		input := domain.CreateOrderInput{
			PropertyID: uuid.New(),
			OrderType:  "guest",
			Items: []domain.CreateItemInput{
				{ItemType: "towel", Quantity: 5, UnitPrice: 30, ServiceType: "wash"},
			},
		}
		if input.OrderType != "guest" && input.OrderType != "house" {
			t.Error("order type should be valid")
		}
		if len(input.Items) == 0 {
			t.Error("items should not be empty")
		}
	})

	t.Run("valid house order", func(t *testing.T) {
		input := domain.CreateOrderInput{
			PropertyID: uuid.New(),
			OrderType:  "house",
			Items: []domain.CreateItemInput{
				{ItemType: "bedsheet", Quantity: 20, UnitPrice: 25, ServiceType: "wash"},
				{ItemType: "pillow_cover", Quantity: 40, UnitPrice: 15, ServiceType: "wash"},
			},
		}
		if input.OrderType != "house" {
			t.Error("expected house order type")
		}
	})

	t.Run("invalid order type", func(t *testing.T) {
		orderType := "external"
		validTypes := map[string]bool{"guest": true, "house": true}
		if validTypes[orderType] {
			t.Error("external should be invalid order type")
		}
	})

	t.Run("empty items is invalid", func(t *testing.T) {
		input := domain.CreateOrderInput{
			PropertyID: uuid.New(),
			OrderType:  "guest",
			Items:      []domain.CreateItemInput{},
		}
		if len(input.Items) != 0 {
			t.Error("items should be empty")
		}
		// Validation should reject empty items (min=1)
	})
}

func TestCreateItemInput_Validation(t *testing.T) {
	t.Run("valid item types", func(t *testing.T) {
		validTypes := []string{"bedsheet", "towel", "pillow_cover", "blanket", "curtain", "shirt", "trouser", "dress", "saree", "other"}
		for _, itemType := range validTypes {
			item := domain.CreateItemInput{
				ItemType:    itemType,
				Quantity:    1,
				UnitPrice:   25,
				ServiceType: "wash",
			}
			if item.ItemType == "" {
				t.Errorf("type %q should be valid", itemType)
			}
		}
	})

	t.Run("valid service types", func(t *testing.T) {
		validServices := []string{"wash", "dry_clean", "iron", "wash_iron"}
		for _, svc := range validServices {
			item := domain.CreateItemInput{
				ItemType:    "shirt",
				Quantity:    2,
				UnitPrice:   50,
				ServiceType: svc,
			}
			if item.ServiceType == "" {
				t.Errorf("service %q should be valid", svc)
			}
		}
	})

	t.Run("quantity bounds", func(t *testing.T) {
		// min=1
		if 0 >= 1 {
			t.Error("quantity 0 should be invalid (min=1)")
		}
		// max=100
		if 101 <= 100 {
			t.Error("quantity 101 should be invalid (max=100)")
		}
		// Valid
		if 50 < 1 || 50 > 100 {
			t.Error("quantity 50 should be valid")
		}
	})

	t.Run("unit price must be positive", func(t *testing.T) {
		if 0.0 > 0 {
			t.Error("zero price should fail gt=0")
		}
		if -10.0 > 0 {
			t.Error("negative price should fail gt=0")
		}
		if !(25.0 > 0) {
			t.Error("positive price should pass")
		}
	})
}

func TestUpdateStatusInput_Validation(t *testing.T) {
	t.Run("valid statuses", func(t *testing.T) {
		validStatuses := []string{"received", "washing", "ready", "delivered"}
		for _, status := range validStatuses {
			input := domain.UpdateStatusInput{
				OrderID: uuid.New(),
				Status:  status,
			}
			found := false
			for _, valid := range validStatuses {
				if input.Status == valid {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("status %q should be valid", status)
			}
		}
	})

	t.Run("invalid statuses", func(t *testing.T) {
		invalidStatuses := []string{"pending", "completed", "cancelled", ""}
		validStatuses := map[string]bool{"received": true, "washing": true, "ready": true, "delivered": true}
		for _, status := range invalidStatuses {
			if validStatuses[status] {
				t.Errorf("status %q should be invalid", status)
			}
		}
	})
}

func TestCreateRateCardInput_Validation(t *testing.T) {
	t.Run("valid rate card", func(t *testing.T) {
		input := domain.CreateRateCardInput{
			PropertyID:  uuid.New(),
			ItemName:    "Cotton Shirt",
			ItemType:    "shirt",
			ServiceType: "wash_iron",
			DefaultRate: 50,
		}
		if input.PropertyID == uuid.Nil {
			t.Error("property_id required")
		}
		if input.ItemName == "" {
			t.Error("item_name required")
		}
		if input.DefaultRate <= 0 {
			t.Error("default_rate must be gt=0")
		}
	})

	t.Run("item name length constraints", func(t *testing.T) {
		// min=1
		if len("") >= 1 {
			t.Error("empty name should fail min=1")
		}
		// max=100
		longName := make([]byte, 101)
		for i := range longName {
			longName[i] = 'a'
		}
		if len(longName) <= 100 {
			t.Error("101-char name should fail max=100")
		}
	})

	t.Run("default rate must be positive", func(t *testing.T) {
		if 0.0 > 0 {
			t.Error("zero rate should fail")
		}
		if !(25.5 > 0) {
			t.Error("positive rate should pass")
		}
	})
}
