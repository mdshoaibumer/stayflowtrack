package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/billing/domain"
)

func TestGetSACCode(t *testing.T) {
	tests := []struct {
		category string
		expected string
	}{
		{"room_charge", "996311"},
		{"food_beverage", "996331"},
		{"laundry", "998112"},
		{"parking", "996742"},
		{"spa", "999723"},
		{"late_checkout", "996311"},
		{"extra_bed", "996311"},
		{"unknown", "999799"},
		{"", "999799"},
		{"minibar", "999799"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			result := getSACCode(tt.category)
			if result != tt.expected {
				t.Errorf("getSACCode(%q) = %q, want %q", tt.category, result, tt.expected)
			}
		})
	}
}

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"shorter than max", "Hello", 10, "Hello"},
		{"exactly max", "Hello", 5, "Hello"},
		{"longer than max", "Hello World!", 8, "Hello..."},
		{"empty string", "", 5, ""},
		{"single char over", "ABCDEF", 5, "AB..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateStr(tt.input, tt.max)
			if result != tt.expected {
				t.Errorf("truncateStr(%q, %d) = %q, want %q", tt.input, tt.max, result, tt.expected)
			}
		})
	}
}

func TestAmountToWords(t *testing.T) {
	tests := []struct {
		amount   int64
		expected string
	}{
		{0, "Zero Rupees"},
		{1, "One Rupees Only"},
		{15, "Fifteen Rupees Only"},
		{99, "Ninety Nine Rupees Only"},
		{100, "One Hundred  Rupees Only"},
		{1000, "One Thousand  Rupees Only"},
		{5000, "Five Thousand  Rupees Only"},
		{12500, "Twelve Thousand Five Hundred  Rupees Only"},
		{100000, "One Lakh  Rupees Only"},
		{250000, "Two Lakh Fifty  Thousand  Rupees Only"},
		{10000000, "One Crore  Rupees Only"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := amountToWords(tt.amount)
			if result != tt.expected {
				t.Errorf("amountToWords(%d) = %q, want %q", tt.amount, result, tt.expected)
			}
		})
	}
}

func TestGeneratePDF(t *testing.T) {
	t.Run("valid invoice generates PDF bytes", func(t *testing.T) {
		invoice := &domain.Invoice{
			ID:              uuid.New(),
			InvoiceNumber:   "INV-2026-001",
			PropertyName:    "StayFlow Apartments",
			PropertyAddress: "123 Main Street, Bangalore",
			PropertyGST:     "29ABCDE1234F1Z5",
			GuestName:       "Raj Sharma",
			GuestPhone:      "9876543210",
			GuestEmail:      "raj@example.com",
			CheckInDate:     time.Now().AddDate(0, 0, -3),
			CheckOutDate:    time.Now(),
			NumNights:       3,
			Subtotal:        decimal.NewFromInt(9000),
			CGSTAmount:      decimal.NewFromInt(540),
			SGSTAmount:      decimal.NewFromInt(540),
			IGSTAmount:      decimal.Zero,
			TotalTax:        decimal.NewFromInt(1080),
			TotalAmount:     decimal.NewFromInt(10080),
			PaidAmount:      decimal.NewFromInt(5000),
			BalanceDue:      decimal.NewFromInt(5080),
			IssuedAt:        time.Now(),
			LineItems: []domain.InvoiceLineItem{
				{
					Description: "Room Charge - Unit 101",
					Category:    "room_charge",
					Quantity:    3,
					UnitPrice:   decimal.NewFromInt(3000),
					TaxRate:     decimal.NewFromInt(12),
					TaxAmount:   decimal.NewFromInt(1080),
					Total:       decimal.NewFromInt(10080),
				},
			},
		}

		result, err := generatePDF(invoice)
		if err != nil {
			t.Fatalf("generatePDF returned error: %v", err)
		}
		if len(result) == 0 {
			t.Error("generatePDF returned empty bytes")
		}
		// PDF files start with %PDF
		if len(result) > 4 && string(result[:4]) != "%PDF" {
			t.Error("generated file does not start with %PDF header")
		}
	})

	t.Run("invoice with no line items", func(t *testing.T) {
		invoice := &domain.Invoice{
			ID:            uuid.New(),
			InvoiceNumber: "INV-2026-002",
			PropertyName:  "Test Property",
			GuestName:     "Test Guest",
			CheckInDate:   time.Now().AddDate(0, 0, -1),
			CheckOutDate:  time.Now(),
			NumNights:     1,
			Subtotal:      decimal.Zero,
			TotalAmount:   decimal.Zero,
			PaidAmount:    decimal.Zero,
			BalanceDue:    decimal.Zero,
			CGSTAmount:    decimal.Zero,
			SGSTAmount:    decimal.Zero,
			IGSTAmount:    decimal.Zero,
			TotalTax:      decimal.Zero,
			IssuedAt:      time.Now(),
			LineItems:     []domain.InvoiceLineItem{},
		}

		result, err := generatePDF(invoice)
		if err != nil {
			t.Fatalf("generatePDF returned error: %v", err)
		}
		if len(result) == 0 {
			t.Error("generatePDF returned empty bytes for empty invoice")
		}
	})

	t.Run("invoice with IGST only (inter-state)", func(t *testing.T) {
		invoice := &domain.Invoice{
			ID:            uuid.New(),
			InvoiceNumber: "INV-2026-003",
			PropertyName:  "Test Property",
			PropertyGST:   "29ABCDE1234F1Z5",
			GuestName:     "Out-of-state Guest",
			GuestAddress:  "Mumbai, Maharashtra",
			CheckInDate:   time.Now().AddDate(0, 0, -2),
			CheckOutDate:  time.Now(),
			NumNights:     2,
			Subtotal:      decimal.NewFromInt(6000),
			CGSTAmount:    decimal.Zero,
			SGSTAmount:    decimal.Zero,
			IGSTAmount:    decimal.NewFromInt(720),
			TotalTax:      decimal.NewFromInt(720),
			TotalAmount:   decimal.NewFromInt(6720),
			PaidAmount:    decimal.NewFromInt(6720),
			BalanceDue:    decimal.Zero,
			IssuedAt:      time.Now(),
			LineItems: []domain.InvoiceLineItem{
				{
					Description: "Room Charge",
					Category:    "room_charge",
					Quantity:    2,
					UnitPrice:   decimal.NewFromInt(3000),
					TaxRate:     decimal.NewFromInt(12),
					TaxAmount:   decimal.NewFromInt(720),
					Total:       decimal.NewFromInt(6720),
				},
			},
		}

		result, err := generatePDF(invoice)
		if err != nil {
			t.Fatalf("generatePDF returned error: %v", err)
		}
		if len(result) == 0 {
			t.Error("generatePDF returned empty bytes")
		}
	})
}
