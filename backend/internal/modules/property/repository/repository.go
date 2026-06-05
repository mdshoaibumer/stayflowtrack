package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stayflow/stayflow-track/internal/modules/property/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateProperty(ctx context.Context, p *domain.Property) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO properties (tenant_id, name, address, city, state, country, pincode, phone, email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, total_units, status, created_at, updated_at`,
		p.TenantID, p.Name, p.Address, p.City, p.State, p.Country, p.Pincode, p.Phone, p.Email,
	).Scan(&p.ID, &p.TotalUnits, &p.Status, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create property: %w", err)
	}
	return nil
}

func (r *Repository) GetPropertyByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Property, error) {
	var p domain.Property
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, address, city, state, country, pincode, phone, email, total_units, status, created_at, updated_at
		 FROM properties WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&p.ID, &p.TenantID, &p.Name, &p.Address, &p.City, &p.State, &p.Country,
		&p.Pincode, &p.Phone, &p.Email, &p.TotalUnits, &p.Status, &p.CreatedAt, &p.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("property", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get property: %w", err)
	}
	return &p, nil
}

func (r *Repository) ListProperties(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Property, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM properties WHERE tenant_id = $1 AND status = 'active'`, tenantID,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count properties: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, address, city, state, country, pincode, phone, email, total_units, status, created_at, updated_at
		 FROM properties WHERE tenant_id = $1 AND status = 'active' ORDER BY name LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list properties: %w", err)
	}
	defer rows.Close()

	var properties []domain.Property
	for rows.Next() {
		var p domain.Property
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Address, &p.City, &p.State, &p.Country,
			&p.Pincode, &p.Phone, &p.Email, &p.TotalUnits, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan property: %w", err)
		}
		properties = append(properties, p)
	}

	if properties == nil {
		properties = []domain.Property{}
	}

	return properties, count, nil
}

func (r *Repository) UpdateProperty(ctx context.Context, p *domain.Property) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE properties SET
			name = COALESCE(NULLIF($3, ''), name),
			address = COALESCE(NULLIF($4, ''), address),
			city = COALESCE(NULLIF($5, ''), city),
			state = COALESCE(NULLIF($6, ''), state),
			country = COALESCE(NULLIF($7, ''), country),
			pincode = COALESCE(NULLIF($8, ''), pincode),
			phone = COALESCE(NULLIF($9, ''), phone),
			email = COALESCE(NULLIF($10, ''), email)
		 WHERE id = $1 AND tenant_id = $2
		 RETURNING id, tenant_id, name, address, city, state, country, pincode, phone, email, total_units, status, created_at, updated_at`,
		p.ID, p.TenantID, p.Name, p.Address, p.City, p.State, p.Country, p.Pincode, p.Phone, p.Email,
	).Scan(&p.ID, &p.TenantID, &p.Name, &p.Address, &p.City, &p.State, &p.Country,
		&p.Pincode, &p.Phone, &p.Email, &p.TotalUnits, &p.Status, &p.CreatedAt, &p.UpdatedAt)

	if err == pgx.ErrNoRows {
		return apperrors.NotFound("property", p.ID.String())
	}
	if err != nil {
		return fmt.Errorf("update property: %w", err)
	}
	return nil
}

func (r *Repository) CreateUnitType(ctx context.Context, ut *domain.UnitType) error {
	amenities, _ := json.Marshal(ut.Amenities)

	err := r.pool.QueryRow(ctx,
		`INSERT INTO unit_types (tenant_id, property_id, name, description, base_rate, max_occupancy, amenities)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		ut.TenantID, ut.PropertyID, ut.Name, ut.Description, ut.BaseRate, ut.MaxOccupancy, amenities,
	).Scan(&ut.ID, &ut.CreatedAt, &ut.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create unit type: %w", err)
	}
	return nil
}

func (r *Repository) ListUnitTypesByProperty(ctx context.Context, propertyID, tenantID uuid.UUID) ([]domain.UnitType, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, property_id, name, description, base_rate, max_occupancy, amenities, created_at, updated_at
		 FROM unit_types WHERE property_id = $1 AND tenant_id = $2 ORDER BY name`,
		propertyID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list unit types: %w", err)
	}
	defer rows.Close()

	var unitTypes []domain.UnitType
	for rows.Next() {
		var ut domain.UnitType
		var amenities []byte
		if err := rows.Scan(&ut.ID, &ut.TenantID, &ut.PropertyID, &ut.Name, &ut.Description,
			&ut.BaseRate, &ut.MaxOccupancy, &amenities, &ut.CreatedAt, &ut.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan unit type: %w", err)
		}
		json.Unmarshal(amenities, &ut.Amenities)
		unitTypes = append(unitTypes, ut)
	}

	if unitTypes == nil {
		unitTypes = []domain.UnitType{}
	}

	return unitTypes, nil
}

func (r *Repository) CreateUnit(ctx context.Context, u *domain.Unit) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO units (tenant_id, property_id, unit_type_id, unit_number, floor, status, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		u.TenantID, u.PropertyID, u.UnitTypeID, u.UnitNumber, u.Floor, u.Status, u.Notes,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create unit: %w", err)
	}

	// Update property total_units count
	_, err = r.pool.Exec(ctx,
		`UPDATE properties SET total_units = (SELECT COUNT(*) FROM units WHERE property_id = $1) WHERE id = $1`,
		u.PropertyID,
	)

	return err
}

func (r *Repository) GetUnitByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Unit, error) {
	var u domain.Unit
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.tenant_id, u.property_id, u.unit_type_id, u.unit_number, u.floor, u.status, u.notes,
		        ut.name, ut.base_rate, u.created_at, u.updated_at
		 FROM units u JOIN unit_types ut ON u.unit_type_id = ut.id
		 WHERE u.id = $1 AND u.tenant_id = $2`, id, tenantID,
	).Scan(&u.ID, &u.TenantID, &u.PropertyID, &u.UnitTypeID, &u.UnitNumber, &u.Floor, &u.Status, &u.Notes,
		&u.UnitTypeName, &u.BaseRate, &u.CreatedAt, &u.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("unit", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get unit: %w", err)
	}
	return &u, nil
}

func (r *Repository) ListUnits(ctx context.Context, propertyID, tenantID uuid.UUID, limit, offset int) ([]domain.Unit, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM units WHERE property_id = $1 AND tenant_id = $2`, propertyID, tenantID,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count units: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.tenant_id, u.property_id, u.unit_type_id, u.unit_number, u.floor, u.status, u.notes,
		        ut.name, ut.base_rate, u.created_at, u.updated_at
		 FROM units u JOIN unit_types ut ON u.unit_type_id = ut.id
		 WHERE u.property_id = $1 AND u.tenant_id = $2
		 ORDER BY u.unit_number LIMIT $3 OFFSET $4`,
		propertyID, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list units: %w", err)
	}
	defer rows.Close()

	var units []domain.Unit
	for rows.Next() {
		var u domain.Unit
		if err := rows.Scan(&u.ID, &u.TenantID, &u.PropertyID, &u.UnitTypeID, &u.UnitNumber, &u.Floor, &u.Status, &u.Notes,
			&u.UnitTypeName, &u.BaseRate, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan unit: %w", err)
		}
		units = append(units, u)
	}

	if units == nil {
		units = []domain.Unit{}
	}

	return units, count, nil
}

func (r *Repository) UpdateUnit(ctx context.Context, u *domain.Unit) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE units SET
			unit_number = COALESCE(NULLIF($3, ''), unit_number),
			floor = COALESCE(NULLIF($4, ''), floor),
			notes = $5
		 WHERE id = $1 AND tenant_id = $2
		 RETURNING id, tenant_id, property_id, unit_type_id, unit_number, floor, status, notes, created_at, updated_at`,
		u.ID, u.TenantID, u.UnitNumber, u.Floor, u.Notes,
	).Scan(&u.ID, &u.TenantID, &u.PropertyID, &u.UnitTypeID, &u.UnitNumber, &u.Floor, &u.Status, &u.Notes,
		&u.CreatedAt, &u.UpdatedAt)

	if err == pgx.ErrNoRows {
		return apperrors.NotFound("unit", u.ID.String())
	}
	if err != nil {
		return fmt.Errorf("update unit: %w", err)
	}
	return nil
}

func (r *Repository) DeleteUnit(ctx context.Context, id, tenantID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM units WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("delete unit: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("unit", id.String())
	}
	return nil
}

func (r *Repository) ChangeUnitStatus(ctx context.Context, id, tenantID uuid.UUID, status domain.UnitStatus) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE units SET status = $3 WHERE id = $1 AND tenant_id = $2`, id, tenantID, string(status),
	)
	if err != nil {
		return fmt.Errorf("change unit status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("unit", id.String())
	}
	return nil
}

type SearchUnitsParams struct {
	PropertyID uuid.UUID
	TenantID   uuid.UUID
	Status     string
	UnitTypeID *uuid.UUID
	Floor      string
	Limit      int
	Offset     int
}

func (r *Repository) SearchUnits(ctx context.Context, params SearchUnitsParams) ([]domain.Unit, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM units u
		 WHERE u.property_id = $1 AND u.tenant_id = $2
		   AND ($3::VARCHAR = '' OR u.status = $3::VARCHAR)
		   AND ($4::UUID IS NULL OR u.unit_type_id = $4::UUID)
		   AND ($5::VARCHAR = '' OR u.floor = $5::VARCHAR)`,
		params.PropertyID, params.TenantID, params.Status, params.UnitTypeID, params.Floor,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count search units: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.tenant_id, u.property_id, u.unit_type_id, u.unit_number, u.floor, u.status, u.notes,
		        ut.name, ut.base_rate, u.created_at, u.updated_at
		 FROM units u JOIN unit_types ut ON u.unit_type_id = ut.id
		 WHERE u.property_id = $1 AND u.tenant_id = $2
		   AND ($3::VARCHAR = '' OR u.status = $3::VARCHAR)
		   AND ($4::UUID IS NULL OR u.unit_type_id = $4::UUID)
		   AND ($5::VARCHAR = '' OR u.floor = $5::VARCHAR)
		 ORDER BY u.unit_number LIMIT $6 OFFSET $7`,
		params.PropertyID, params.TenantID, params.Status, params.UnitTypeID, params.Floor, params.Limit, params.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("search units: %w", err)
	}
	defer rows.Close()

	var units []domain.Unit
	for rows.Next() {
		var u domain.Unit
		if err := rows.Scan(&u.ID, &u.TenantID, &u.PropertyID, &u.UnitTypeID, &u.UnitNumber, &u.Floor, &u.Status, &u.Notes,
			&u.UnitTypeName, &u.BaseRate, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan unit: %w", err)
		}
		units = append(units, u)
	}

	if units == nil {
		units = []domain.Unit{}
	}

	return units, count, nil
}
