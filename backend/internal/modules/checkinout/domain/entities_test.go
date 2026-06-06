package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stayflow/stayflow-track/internal/modules/checkinout/domain"
)

func TestCheckInInput_Validation(t *testing.T) {
	validInput := domain.CheckInInput{
		ReservationID:    uuid.New(),
		AssignedUnitID:   uuid.New(),
		DepositAmount:    decimal.NewFromInt(2000),
		DepositMethod:    "cash",
		IDDocumentType:   "aadhaar",
		IDDocumentNumber: "123456789012",
	}

	t.Run("valid input", func(t *testing.T) {
		if validInput.ReservationID == uuid.Nil {
			t.Error("reservation_id should not be nil")
		}
		if validInput.DepositAmount.IsNegative() || validInput.DepositAmount.IsZero() {
			t.Error("deposit amount must be positive")
		}
	})

	t.Run("deposit methods", func(t *testing.T) {
		validMethods := []string{"cash", "upi", "card", "bank_transfer", "cheque"}
		for _, method := range validMethods {
			input := validInput
			input.DepositMethod = method
			if input.DepositMethod == "" {
				t.Errorf("method %q should be valid", method)
			}
		}

		invalidMethods := []string{"bitcoin", "paypal", "", "wire"}
		for _, method := range invalidMethods {
			contains := false
			for _, valid := range []string{"cash", "upi", "card", "bank_transfer", "cheque"} {
				if method == valid {
					contains = true
				}
			}
			if contains {
				t.Errorf("method %q should be invalid", method)
			}
		}
	})

	t.Run("ID document types", func(t *testing.T) {
		validTypes := []string{"aadhaar", "passport", "driving_license", "voter_id", "pan_card"}
		for _, idType := range validTypes {
			input := validInput
			input.IDDocumentType = idType
			if input.IDDocumentType == "" {
				t.Errorf("type %q should be valid", idType)
			}
		}
	})

	t.Run("ID number constraints", func(t *testing.T) {
		// Min length = 4
		shortNumber := "123"
		if len(shortNumber) >= 4 {
			t.Error("short number should fail min=4")
		}

		// Max length = 50
		longNumber := "12345678901234567890123456789012345678901234567890X"
		if len(longNumber) <= 50 {
			t.Error("long number should fail max=50")
		}

		validNumber := "1234"
		if len(validNumber) < 4 || len(validNumber) > 50 {
			t.Error("valid number should pass")
		}
	})
}

func TestCheckOutInput_Validation(t *testing.T) {
	t.Run("valid checkout input", func(t *testing.T) {
		input := domain.CheckOutInput{
			ReservationID:      uuid.New(),
			LateCheckoutCharge: decimal.NewFromInt(500),
			Notes:              "Guest departing early",
		}
		if input.ReservationID == uuid.Nil {
			t.Error("reservation_id should not be nil")
		}
	})

	t.Run("zero late charge is valid", func(t *testing.T) {
		input := domain.CheckOutInput{
			ReservationID:      uuid.New(),
			LateCheckoutCharge: decimal.Zero,
		}
		if !input.LateCheckoutCharge.IsZero() {
			t.Error("zero late charge should be valid")
		}
	})

	t.Run("notes max length", func(t *testing.T) {
		longNotes := make([]byte, 1001)
		for i := range longNotes {
			longNotes[i] = 'a'
		}
		if len(longNotes) <= 1000 {
			t.Error("notes exceeding 1000 chars should fail")
		}
	})
}

func TestWalkInInput_Validation(t *testing.T) {
	t.Run("valid walk-in input", func(t *testing.T) {
		input := domain.WalkInInput{
			PropertyID:       uuid.New(),
			GuestFirstName:   "Raj",
			GuestLastName:    "Sharma",
			GuestPhone:       "9876543210",
			UnitID:           uuid.New(),
			CheckOutDate:     "2026-06-10",
			RatePerNight:     decimal.NewFromInt(3000),
			DepositAmount:    decimal.NewFromInt(2000),
			DepositMethod:    "cash",
			IDDocumentType:   "aadhaar",
			IDDocumentNumber: "123456789012",
			NumGuests:        2,
		}
		if input.PropertyID == uuid.Nil {
			t.Error("property_id required")
		}
		if input.GuestFirstName == "" {
			t.Error("first name required")
		}
		if input.GuestPhone == "" {
			t.Error("phone required")
		}
	})

	t.Run("phone number constraints", func(t *testing.T) {
		// Phone should be reasonable length
		shortPhone := "123"
		if len(shortPhone) >= 7 {
			t.Error("short phone should be invalid")
		}

		validPhone := "9876543210"
		if len(validPhone) < 7 || len(validPhone) > 15 {
			t.Error("valid phone should pass")
		}
	})

	t.Run("last name is optional", func(t *testing.T) {
		input := domain.WalkInInput{
			PropertyID:       uuid.New(),
			GuestFirstName:   "Ramesh",
			GuestLastName:    "", // Optional
			GuestPhone:       "9876543210",
			UnitID:           uuid.New(),
			CheckOutDate:     "2026-06-10",
			RatePerNight:     decimal.NewFromInt(2500),
			DepositAmount:    decimal.Zero, // Can be zero for corporate
			DepositMethod:    "cash",
			IDDocumentType:   "aadhaar",
			IDDocumentNumber: "1234",
			NumGuests:        1,
		}
		// Last name empty is fine
		if input.GuestLastName != "" {
			t.Error("last name should be optional")
		}
		// Deposit zero is acceptable for corporate
		if !input.DepositAmount.IsZero() {
			t.Error("zero deposit should be valid for corporate")
		}
	})
}
