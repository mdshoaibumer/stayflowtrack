package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/stayflow/stayflow-track/internal/modules/property/domain"
)

// PropertyReader defines the minimal interface needed from the property module.
// This decouples the reservation service from the property repository implementation.
type PropertyReader interface {
	GetUnitByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Unit, error)
}
