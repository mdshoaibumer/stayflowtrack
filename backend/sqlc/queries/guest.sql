-- name: CreateGuest :one
INSERT INTO guests (
    tenant_id, first_name, last_name, email, phone,
    address, city, state, country, pincode,
    nationality, date_of_birth, aadhaar_number, passport_number, notes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetGuestByID :one
SELECT * FROM guests WHERE id = $1 AND tenant_id = $2;

-- name: UpdateGuest :one
UPDATE guests SET
    first_name = COALESCE(NULLIF($3, ''), first_name),
    last_name = COALESCE(NULLIF($4, ''), last_name),
    email = COALESCE(NULLIF($5, ''), email),
    phone = COALESCE(NULLIF($6, ''), phone),
    address = COALESCE(NULLIF($7, ''), address),
    city = COALESCE(NULLIF($8, ''), city),
    state = COALESCE(NULLIF($9, ''), state),
    country = COALESCE(NULLIF($10, ''), country),
    pincode = COALESCE(NULLIF($11, ''), pincode),
    nationality = COALESCE(NULLIF($12, ''), nationality),
    date_of_birth = COALESCE($13, date_of_birth),
    aadhaar_number = COALESCE(NULLIF($14, ''), aadhaar_number),
    passport_number = COALESCE(NULLIF($15, ''), passport_number),
    notes = $16
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: ListGuests :many
SELECT * FROM guests
WHERE tenant_id = $1
ORDER BY last_name, first_name
LIMIT $2 OFFSET $3;

-- name: CountGuests :one
SELECT COUNT(*) FROM guests WHERE tenant_id = $1;

-- name: SearchGuests :many
SELECT * FROM guests
WHERE tenant_id = $1
  AND (
    $2::VARCHAR = ''
    OR first_name ILIKE '%' || $2 || '%'
    OR last_name ILIKE '%' || $2 || '%'
    OR phone ILIKE '%' || $2 || '%'
    OR email ILIKE '%' || $2 || '%'
    OR aadhaar_number = $2
    OR passport_number = $2
  )
ORDER BY last_name, first_name
LIMIT $3 OFFSET $4;

-- name: CountSearchGuests :one
SELECT COUNT(*) FROM guests
WHERE tenant_id = $1
  AND (
    $2::VARCHAR = ''
    OR first_name ILIKE '%' || $2 || '%'
    OR last_name ILIKE '%' || $2 || '%'
    OR phone ILIKE '%' || $2 || '%'
    OR email ILIKE '%' || $2 || '%'
    OR aadhaar_number = $2
    OR passport_number = $2
  );

-- name: IncrementGuestStays :exec
UPDATE guests SET total_stays = total_stays + 1 WHERE id = $1;

-- name: CreateGuestDocument :one
INSERT INTO guest_documents (tenant_id, guest_id, document_type, file_key, file_name, file_size, content_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListGuestDocuments :many
SELECT * FROM guest_documents WHERE guest_id = $1 AND tenant_id = $2 ORDER BY uploaded_at DESC;

-- name: DeleteGuestDocument :exec
DELETE FROM guest_documents WHERE id = $1 AND tenant_id = $2;
