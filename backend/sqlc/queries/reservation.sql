-- name: CreateReservation :one
INSERT INTO reservations (
    tenant_id, property_id, unit_id, guest_id,
    booking_source, status, check_in_date, check_out_date,
    num_guests, rate_per_night, total_amount, notes, external_booking_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetReservationByID :one
SELECT r.*,
    g.first_name as guest_first_name, g.last_name as guest_last_name, g.phone as guest_phone,
    u.unit_number, ut.name as unit_type_name, p.name as property_name
FROM reservations r
JOIN guests g ON r.guest_id = g.id
JOIN units u ON r.unit_id = u.id
JOIN unit_types ut ON u.unit_type_id = ut.id
JOIN properties p ON r.property_id = p.id
WHERE r.id = $1 AND r.tenant_id = $2;

-- name: ListReservations :many
SELECT r.*,
    g.first_name as guest_first_name, g.last_name as guest_last_name, g.phone as guest_phone,
    u.unit_number, ut.name as unit_type_name
FROM reservations r
JOIN guests g ON r.guest_id = g.id
JOIN units u ON r.unit_id = u.id
JOIN unit_types ut ON u.unit_type_id = ut.id
WHERE r.tenant_id = $1
  AND ($2::UUID IS NULL OR r.property_id = $2::UUID)
  AND ($3::VARCHAR = '' OR r.status = $3::VARCHAR)
ORDER BY r.check_in_date DESC
LIMIT $4 OFFSET $5;

-- name: CountReservations :one
SELECT COUNT(*) FROM reservations
WHERE tenant_id = $1
  AND ($2::UUID IS NULL OR property_id = $2::UUID)
  AND ($3::VARCHAR = '' OR status = $3::VARCHAR);

-- name: UpdateReservation :one
UPDATE reservations SET
    check_in_date = COALESCE($3, check_in_date),
    check_out_date = COALESCE($4, check_out_date),
    num_guests = COALESCE($5, num_guests),
    rate_per_night = COALESCE($6, rate_per_night),
    total_amount = COALESCE($7, total_amount),
    notes = $8
WHERE id = $1 AND tenant_id = $2 AND status IN ('pending', 'confirmed')
RETURNING *;

-- name: UpdateReservationStatus :exec
UPDATE reservations SET status = $3 WHERE id = $1 AND tenant_id = $2;

-- name: CancelReservation :exec
UPDATE reservations SET
    status = 'cancelled',
    cancellation_reason = $3,
    cancelled_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND status IN ('pending', 'confirmed');

-- name: CheckIn :exec
UPDATE reservations SET
    status = 'checked_in',
    actual_check_in = NOW()
WHERE id = $1 AND tenant_id = $2 AND status = 'confirmed';

-- name: CheckOut :exec
UPDATE reservations SET
    status = 'checked_out',
    actual_check_out = NOW()
WHERE id = $1 AND tenant_id = $2 AND status = 'checked_in';

-- name: CheckConflict :one
SELECT check_reservation_conflict($1, $2, $3, $4);

-- name: GetAvailableUnits :many
SELECT u.*, ut.name as unit_type_name, ut.base_rate
FROM units u
JOIN unit_types ut ON u.unit_type_id = ut.id
WHERE u.property_id = $1
  AND u.tenant_id = $2
  AND u.status = 'available'
  AND u.id NOT IN (
      SELECT unit_id FROM reservations
      WHERE property_id = $1
        AND status NOT IN ('cancelled', 'checked_out')
        AND check_in_date < $4
        AND check_out_date > $3
  )
ORDER BY u.unit_number;

-- name: GetGuestReservations :many
SELECT r.*, u.unit_number, ut.name as unit_type_name, p.name as property_name
FROM reservations r
JOIN units u ON r.unit_id = u.id
JOIN unit_types ut ON u.unit_type_id = ut.id
JOIN properties p ON r.property_id = p.id
WHERE r.guest_id = $1 AND r.tenant_id = $2
ORDER BY r.check_in_date DESC
LIMIT $3 OFFSET $4;

-- name: CreateAuditLog :exec
INSERT INTO audit_logs (tenant_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
