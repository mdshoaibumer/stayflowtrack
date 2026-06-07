package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestFolioStatus_Constants(t *testing.T) {
	if FolioOpen != "open" {
		t.Errorf("expected 'open', got %q", FolioOpen)
	}
	if FolioClosed != "closed" {
		t.Errorf("expected 'closed', got %q", FolioClosed)
	}
	if FolioVoid != "void" {
		t.Errorf("expected 'void', got %q", FolioVoid)
	}
}

func TestLineItemCategory_Constants(t *testing.T) {
	categories := map[LineItemCategory]string{
		CategoryRoomCharge:   "room_charge",
		CategoryDeposit:      "deposit",
		CategoryFoodBeverage: "food_beverage",
		CategoryLaundry:      "laundry",
		CategoryMinibar:      "minibar",
		CategoryParking:      "parking",
		CategorySpa:          "spa",
		CategoryDamage:       "damage",
		CategoryLateCheckout: "late_checkout",
		CategoryExtraBed:     "extra_bed",
		CategoryOther:        "other",
	}

	for cat, expected := range categories {
		if string(cat) != expected {
			t.Errorf("category %q != expected %q", cat, expected)
		}
	}
}

func TestInvoiceStatus_Constants(t *testing.T) {
	statuses := map[InvoiceStatus]string{
		InvoiceIssued:        "issued",
		InvoicePaid:          "paid",
		InvoicePartiallyPaid: "partially_paid",
		InvoiceVoid:          "void",
		InvoiceRefunded:      "refunded",
	}

	for s, expected := range statuses {
		if string(s) != expected {
			t.Errorf("status %q != expected %q", s, expected)
		}
	}
}

func TestFolio_BalanceCalculation(t *testing.T) {
	folio := Folio{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Status:      FolioOpen,
		Subtotal:    decimal.NewFromInt(9000),
		TaxTotal:    decimal.NewFromInt(1080),
		TotalAmount: decimal.NewFromInt(10080),
		PaidAmount:  decimal.NewFromInt(5000),
		Balance:     decimal.NewFromInt(5080),
	}

	expectedBalance := folio.TotalAmount.Sub(folio.PaidAmount)
	if !folio.Balance.Equal(expectedBalance) {
		t.Errorf("balance %s != total-paid %s", folio.Balance.String(), expectedBalance.String())
	}
}

func TestLineItem_TaxCalculation(t *testing.T) {
	tests := []struct {
		name      string
		quantity  int
		unitPrice decimal.Decimal
		taxRate   decimal.Decimal
		expAmount decimal.Decimal
		expTax    decimal.Decimal
		expTotal  decimal.Decimal
	}{
		{
			name:      "room charge 12% GST",
			quantity:  3,
			unitPrice: decimal.NewFromInt(3000),
			taxRate:   decimal.NewFromInt(12),
			expAmount: decimal.NewFromInt(9000),
			expTax:    decimal.NewFromInt(1080),
			expTotal:  decimal.NewFromInt(10080),
		},
		{
			name:      "food 5% GST",
			quantity:  2,
			unitPrice: decimal.NewFromInt(500),
			taxRate:   decimal.NewFromInt(5),
			expAmount: decimal.NewFromInt(1000),
			expTax:    decimal.NewFromInt(50),
			expTotal:  decimal.NewFromInt(1050),
		},
		{
			name:      "spa 18% GST",
			quantity:  1,
			unitPrice: decimal.NewFromInt(2000),
			taxRate:   decimal.NewFromInt(18),
			expAmount: decimal.NewFromInt(2000),
			expTax:    decimal.NewFromInt(360),
			expTotal:  decimal.NewFromInt(2360),
		},
		{
			name:      "zero tax",
			quantity:  1,
			unitPrice: decimal.NewFromInt(100),
			taxRate:   decimal.Zero,
			expAmount: decimal.NewFromInt(100),
			expTax:    decimal.Zero,
			expTotal:  decimal.NewFromInt(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount := tt.unitPrice.Mul(decimal.NewFromInt(int64(tt.quantity)))
			taxAmount := amount.Mul(tt.taxRate).Div(decimal.NewFromInt(100))
			total := amount.Add(taxAmount)

			if !amount.Equal(tt.expAmount) {
				t.Errorf("amount: got %s, want %s", amount.String(), tt.expAmount.String())
			}
			if !taxAmount.Equal(tt.expTax) {
				t.Errorf("tax: got %s, want %s", taxAmount.String(), tt.expTax.String())
			}
			if !total.Equal(tt.expTotal) {
				t.Errorf("total: got %s, want %s", total.String(), tt.expTotal.String())
			}
		})
	}
}

func TestLineItem_VoidState(t *testing.T) {
	item := LineItem{
		ID:         uuid.New(),
		TenantID:   uuid.New(),
		FolioID:    uuid.New(),
		Category:   CategoryRoomCharge,
		Quantity:   1,
		UnitPrice:  decimal.NewFromInt(3000),
		IsVoid:     true,
		VoidReason: "Incorrect charge",
	}

	if !item.IsVoid {
		t.Error("expected item to be void")
	}
	if item.VoidReason == "" {
		t.Error("void items should have a reason")
	}
}

func TestInvoice_GSTSplit(t *testing.T) {
	// For intra-state: CGST + SGST = Total Tax, IGST = 0
	invoice := Invoice{
		ID:         uuid.New(),
		Subtotal:   decimal.NewFromInt(10000),
		CGSTAmount: decimal.NewFromInt(600),
		SGSTAmount: decimal.NewFromInt(600),
		IGSTAmount: decimal.Zero,
		TotalTax:   decimal.NewFromInt(1200),
	}

	gstSum := invoice.CGSTAmount.Add(invoice.SGSTAmount).Add(invoice.IGSTAmount)
	if !gstSum.Equal(invoice.TotalTax) {
		t.Errorf("CGST+SGST+IGST=%s != TotalTax=%s", gstSum.String(), invoice.TotalTax.String())
	}
}

func TestInvoice_InterState_IGST(t *testing.T) {
	// For inter-state: IGST = Total Tax, CGST = SGST = 0
	invoice := Invoice{
		ID:         uuid.New(),
		Subtotal:   decimal.NewFromInt(10000),
		CGSTAmount: decimal.Zero,
		SGSTAmount: decimal.Zero,
		IGSTAmount: decimal.NewFromInt(1200),
		TotalTax:   decimal.NewFromInt(1200),
	}

	if !invoice.CGSTAmount.IsZero() {
		t.Error("CGST should be zero for inter-state")
	}
	if !invoice.SGSTAmount.IsZero() {
		t.Error("SGST should be zero for inter-state")
	}
	if !invoice.IGSTAmount.Equal(invoice.TotalTax) {
		t.Errorf("IGST %s should equal TotalTax %s for inter-state", invoice.IGSTAmount.String(), invoice.TotalTax.String())
	}
}
