package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Property struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	City       string    `json:"city"`
	State      string    `json:"state"`
	Country    string    `json:"country"`
	Pincode    string    `json:"pincode,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Email      string    `json:"email,omitempty"`
	TotalUnits int       `json:"total_units"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UnitType struct {
	ID           uuid.UUID       `json:"id"`
	TenantID     uuid.UUID       `json:"tenant_id"`
	PropertyID   uuid.UUID       `json:"property_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	BaseRate     decimal.Decimal `json:"base_rate"`
	MaxOccupancy int             `json:"max_occupancy"`
	Amenities    []string        `json:"amenities,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type UnitStatus string

const (
	UnitStatusAvailable   UnitStatus = "available"
	UnitStatusReserved    UnitStatus = "reserved"
	UnitStatusOccupied    UnitStatus = "occupied"
	UnitStatusCleaning    UnitStatus = "cleaning"
	UnitStatusMaintenance UnitStatus = "maintenance"
)

func (s UnitStatus) IsValid() bool {
	switch s {
	case UnitStatusAvailable, UnitStatusReserved, UnitStatusOccupied,
		UnitStatusCleaning, UnitStatusMaintenance:
		return true
	}
	return false
}

type Unit struct {
	ID           uuid.UUID       `json:"id"`
	TenantID     uuid.UUID       `json:"tenant_id"`
	PropertyID   uuid.UUID       `json:"property_id"`
	UnitTypeID   uuid.UUID       `json:"unit_type_id"`
	UnitNumber   string          `json:"unit_number"`
	Floor        string          `json:"floor,omitempty"`
	Status       UnitStatus      `json:"status"`
	Notes        string          `json:"notes,omitempty"`
	UnitTypeName string          `json:"unit_type_name,omitempty"`
	BaseRate     decimal.Decimal `json:"base_rate,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
