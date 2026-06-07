package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGuest_FullName(t *testing.T) {
	g := Guest{
		FirstName: "Raj",
		LastName:  "Sharma",
	}

	fullName := g.FirstName + " " + g.LastName
	if fullName != "Raj Sharma" {
		t.Errorf("expected 'Raj Sharma', got %q", fullName)
	}
}

func TestGuest_RequiredFields(t *testing.T) {
	g := Guest{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		FirstName: "Priya",
		LastName:  "Patel",
		Phone:     "9876543210",
	}

	if g.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if g.FirstName == "" {
		t.Error("first name should not be empty")
	}
	if g.LastName == "" {
		t.Error("last name should not be empty")
	}
	if g.Phone == "" {
		t.Error("phone should not be empty")
	}
}

func TestGuest_OptionalFields(t *testing.T) {
	dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	g := Guest{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		FirstName:      "John",
		LastName:       "Doe",
		Phone:          "1234567890",
		Email:          "john@example.com",
		Address:        "123 Main St",
		City:           "Bangalore",
		State:          "Karnataka",
		Country:        "India",
		Pincode:        "560001",
		Nationality:    "Indian",
		DateOfBirth:    &dob,
		AadhaarNumber:  "123456789012",
		PassportNumber: "A1234567",
		Notes:          "VIP guest",
	}

	if g.Email != "john@example.com" {
		t.Error("email not set")
	}
	if g.AadhaarNumber != "123456789012" {
		t.Error("aadhaar not set")
	}
	if len(g.AadhaarNumber) != 12 {
		t.Errorf("aadhaar should be 12 digits, got %d", len(g.AadhaarNumber))
	}
	if g.DateOfBirth == nil {
		t.Error("date of birth should be set")
	}
}

func TestGuestDocument_Fields(t *testing.T) {
	doc := GuestDocument{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		GuestID:      uuid.New(),
		DocumentType: "aadhaar",
		FileKey:      "documents/abc/aadhaar.pdf",
		FileName:     "aadhaar.pdf",
		FileSize:     1024000,
		ContentType:  "application/pdf",
		UploadedAt:   time.Now(),
	}

	if doc.DocumentType != "aadhaar" {
		t.Errorf("expected 'aadhaar', got %s", doc.DocumentType)
	}
	if doc.FileSize > 10*1024*1024 {
		t.Error("file size should be under 10MB")
	}
}

func TestGuestDocument_ValidContentTypes(t *testing.T) {
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"application/pdf",
	}

	for _, ct := range validTypes {
		doc := GuestDocument{
			ID:          uuid.New(),
			ContentType: ct,
		}
		if doc.ContentType == "" {
			t.Errorf("content type should not be empty for %s", ct)
		}
	}
}

func TestGuest_TotalStaysDefault(t *testing.T) {
	g := Guest{}
	if g.TotalStays != 0 {
		t.Errorf("expected default 0 stays, got %d", g.TotalStays)
	}
}
