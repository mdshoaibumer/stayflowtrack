package domain

import (
	"time"

	"github.com/google/uuid"
)

type Guest struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Email          string     `json:"email,omitempty"`
	Phone          string     `json:"phone"`
	Address        string     `json:"address,omitempty"`
	City           string     `json:"city,omitempty"`
	State          string     `json:"state,omitempty"`
	Country        string     `json:"country,omitempty"`
	Pincode        string     `json:"pincode,omitempty"`
	Nationality    string     `json:"nationality,omitempty"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	AadhaarNumber  string     `json:"aadhaar_number,omitempty"`
	PassportNumber string     `json:"passport_number,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	TotalStays     int        `json:"total_stays"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type GuestDocument struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	GuestID      uuid.UUID `json:"guest_id"`
	DocumentType string    `json:"document_type"`
	FileKey      string    `json:"file_key"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	ContentType  string    `json:"content_type"`
	UploadedAt   time.Time `json:"uploaded_at"`
	DownloadURL  string    `json:"download_url,omitempty"`
}
