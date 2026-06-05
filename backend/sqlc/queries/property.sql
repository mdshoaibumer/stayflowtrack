-- name: CreateProperty :one
INSERT INTO properties (tenant_id, name, address, city, state, country, pincode, phone, email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetPropertyByID :one
SELECT * FROM properties WHERE id = $1 AND tenant_id = $2;

-- name: ListProperties :many
SELECT * FROM properties
WHERE tenant_id = $1 AND status = 'active'
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: CountProperties :one
SELECT COUNT(*) FROM properties WHERE tenant_id = $1 AND status = 'active';

-- name: UpdateProperty :one
UPDATE properties SET
    name = COALESCE(NULLIF($3, ''), name),
    address = COALESCE(NULLIF($4, ''), address),
    city = COALESCE(NULLIF($5, ''), city),
    state = COALESCE(NULLIF($6, ''), state),
    country = COALESCE(NULLIF($7, ''), country),
    pincode = COALESCE(NULLIF($8, ''), pincode),
    phone = COALESCE(NULLIF($9, ''), phone),
    email = COALESCE(NULLIF($10, ''), email)
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: UpdatePropertyTotalUnits :exec
UPDATE properties SET total_units = (
    SELECT COUNT(*) FROM units WHERE property_id = $1
) WHERE id = $1;

-- name: CreateUnitType :one
INSERT INTO unit_types (tenant_id, property_id, name, description, base_rate, max_occupancy, amenities)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUnitTypeByID :one
SELECT * FROM unit_types WHERE id = $1 AND tenant_id = $2;

-- name: ListUnitTypesByProperty :many
SELECT * FROM unit_types WHERE property_id = $1 AND tenant_id = $2 ORDER BY name;

-- name: CreateUnit :one
INSERT INTO units (tenant_id, property_id, unit_type_id, unit_number, floor, status, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUnitByID :one
SELECT u.*, ut.name as unit_type_name, ut.base_rate
FROM units u
JOIN unit_types ut ON u.unit_type_id = ut.id
WHERE u.id = $1 AND u.tenant_id = $2;

-- name: ListUnitsByProperty :many
SELECT u.*, ut.name as unit_type_name, ut.base_rate
FROM units u
JOIN unit_types ut ON u.unit_type_id = ut.id
WHERE u.property_id = $1 AND u.tenant_id = $2
ORDER BY u.unit_number
LIMIT $3 OFFSET $4;

-- name: CountUnitsByProperty :one
SELECT COUNT(*) FROM units WHERE property_id = $1 AND tenant_id = $2;

-- name: UpdateUnit :one
UPDATE units SET
    unit_number = COALESCE(NULLIF($3, ''), unit_number),
    floor = COALESCE(NULLIF($4, ''), floor),
    notes = $5
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: DeleteUnit :exec
DELETE FROM units WHERE id = $1 AND tenant_id = $2;

-- name: ChangeUnitStatus :exec
UPDATE units SET status = $3 WHERE id = $1 AND tenant_id = $2;

-- name: SearchUnits :many
SELECT u.*, ut.name as unit_type_name, ut.base_rate
FROM units u
JOIN unit_types ut ON u.unit_type_id = ut.id
WHERE u.property_id = $1
  AND u.tenant_id = $2
  AND ($3::VARCHAR = '' OR u.status = $3::VARCHAR)
  AND ($4::UUID IS NULL OR u.unit_type_id = $4::UUID)
  AND ($5::VARCHAR = '' OR u.floor = $5::VARCHAR)
ORDER BY u.unit_number
LIMIT $6 OFFSET $7;

-- name: CountSearchUnits :one
SELECT COUNT(*)
FROM units u
WHERE u.property_id = $1
  AND u.tenant_id = $2
  AND ($3::VARCHAR = '' OR u.status = $3::VARCHAR)
  AND ($4::UUID IS NULL OR u.unit_type_id = $4::UUID)
  AND ($5::VARCHAR = '' OR u.floor = $5::VARCHAR);
