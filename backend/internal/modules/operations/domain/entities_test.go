package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestDepositStatus_Constants(t *testing.T) {
	if DepositHeld != "held" {
		t.Errorf("expected 'held', got %q", DepositHeld)
	}
	if DepositApplied != "applied" {
		t.Errorf("expected 'applied', got %q", DepositApplied)
	}
	if DepositRefunded != "refunded" {
		t.Errorf("expected 'refunded', got %q", DepositRefunded)
	}
	if DepositForfeited != "forfeited" {
		t.Errorf("expected 'forfeited', got %q", DepositForfeited)
	}
}

func TestNoShowInput_Validation(t *testing.T) {
	input := NoShowInput{
		ReservationID: uuid.New(),
		ChargeAmount:  decimal.NewFromInt(500),
	}

	if input.ReservationID == uuid.Nil {
		t.Error("expected non-nil reservation ID")
	}
	if !input.ChargeAmount.Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected charge 500, got %s", input.ChargeAmount.String())
	}
}

func TestExtendStayInput_Fields(t *testing.T) {
	input := ExtendStayInput{
		ReservationID: uuid.New(),
		NewCheckOut:   "2026-06-15",
		RatePerNight:  decimal.NewFromInt(3000),
		Reason:        "Guest extended due to meeting",
	}

	if input.NewCheckOut != "2026-06-15" {
		t.Errorf("expected '2026-06-15', got %s", input.NewCheckOut)
	}
	if !input.RatePerNight.Equal(decimal.NewFromInt(3000)) {
		t.Errorf("expected rate 3000, got %s", input.RatePerNight.String())
	}
}

func TestRoomMoveInput_Fields(t *testing.T) {
	input := RoomMoveInput{
		ReservationID: uuid.New(),
		ToUnitID:      uuid.New(),
		Reason:        "Guest requested upgrade",
		RateChange:    decimal.NewFromInt(500),
	}

	if input.Reason != "Guest requested upgrade" {
		t.Errorf("unexpected reason: %s", input.Reason)
	}
	if input.ToUnitID == uuid.Nil {
		t.Error("expected non-nil ToUnitID")
	}
}

func TestMaintenanceBlockInput_BlockTypes(t *testing.T) {
	validTypes := []string{"maintenance", "renovation", "deep_cleaning", "owner_use", "other"}

	for _, bt := range validTypes {
		input := MaintenanceBlockInput{
			PropertyID: uuid.New(),
			UnitID:     uuid.New(),
			Reason:     "Test reason",
			BlockType:  bt,
			StartDate:  "2026-06-10",
			EndDate:    "2026-06-12",
		}
		if input.BlockType != bt {
			t.Errorf("expected block type %q, got %q", bt, input.BlockType)
		}
	}
}

func TestRefundDepositInput_Fields(t *testing.T) {
	input := RefundDepositInput{
		DepositID:       uuid.New(),
		RefundAmount:    decimal.NewFromInt(2000),
		RefundMethod:    "upi",
		RefundReference: "UPI123456",
		Notes:           "Refund at checkout",
	}

	if !input.RefundAmount.Equal(decimal.NewFromInt(2000)) {
		t.Errorf("expected refund 2000, got %s", input.RefundAmount.String())
	}
	if input.RefundMethod != "upi" {
		t.Errorf("expected method 'upi', got %s", input.RefundMethod)
	}
}

func TestStayExtension_AdditionalAmount(t *testing.T) {
	ext := StayExtension{
		ID:               uuid.New(),
		TenantID:         uuid.New(),
		ReservationID:    uuid.New(),
		AdditionalNights: 2,
		RatePerNight:     decimal.NewFromInt(3000),
		AdditionalAmount: decimal.NewFromInt(6000),
	}

	expected := ext.RatePerNight.Mul(decimal.NewFromInt(int64(ext.AdditionalNights)))
	if !ext.AdditionalAmount.Equal(expected) {
		t.Errorf("additional amount %s != rate * nights %s", ext.AdditionalAmount.String(), expected.String())
	}
}

func TestCorporateAccount_Fields(t *testing.T) {
	account := CorporateAccount{
		ID:                     uuid.New(),
		TenantID:               uuid.New(),
		CompanyName:            "Acme Corp",
		CreditLimit:            decimal.NewFromInt(100000),
		PaymentTermsDays:       30,
		NegotiatedRateDiscount: decimal.NewFromFloat(0.15),
		IsActive:               true,
	}

	if account.CompanyName != "Acme Corp" {
		t.Errorf("unexpected company name: %s", account.CompanyName)
	}
	if !account.CreditLimit.Equal(decimal.NewFromInt(100000)) {
		t.Errorf("expected credit limit 100000, got %s", account.CreditLimit.String())
	}
	if account.PaymentTermsDays != 30 {
		t.Errorf("expected 30 days, got %d", account.PaymentTermsDays)
	}
}

func TestRoomMove_RateChangeCanBeNegative(t *testing.T) {
	move := RoomMove{
		ID:            uuid.New(),
		ReservationID: uuid.New(),
		FromUnitID:    uuid.New(),
		ToUnitID:      uuid.New(),
		Reason:        "Downgrade requested",
		RateChange:    decimal.NewFromInt(-500),
	}

	if !move.RateChange.IsNegative() {
		t.Error("expected negative rate change for downgrade")
	}
}
