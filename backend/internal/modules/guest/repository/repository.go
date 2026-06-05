package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stayflow/stayflow-track/internal/modules/guest/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateGuest(ctx context.Context, g *domain.Guest) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO guests (
			tenant_id, first_name, last_name, email, phone,
			address, city, state, country, pincode,
			nationality, date_of_birth, aadhaar_number, passport_number, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, total_stays, created_at, updated_at`,
		g.TenantID, g.FirstName, g.LastName, g.Email, g.Phone,
		g.Address, g.City, g.State, g.Country, g.Pincode,
		g.Nationality, g.DateOfBirth, g.AadhaarNumber, g.PassportNumber, g.Notes,
	).Scan(&g.ID, &g.TotalStays, &g.CreatedAt, &g.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create guest: %w", err)
	}
	return nil
}

func (r *Repository) GetGuestByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Guest, error) {
	var g domain.Guest
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, first_name, last_name, email, phone,
		        address, city, state, country, pincode, nationality,
		        date_of_birth, aadhaar_number, passport_number, notes,
		        total_stays, created_at, updated_at
		 FROM guests WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&g.ID, &g.TenantID, &g.FirstName, &g.LastName, &g.Email, &g.Phone,
		&g.Address, &g.City, &g.State, &g.Country, &g.Pincode, &g.Nationality,
		&g.DateOfBirth, &g.AadhaarNumber, &g.PassportNumber, &g.Notes,
		&g.TotalStays, &g.CreatedAt, &g.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("guest", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get guest: %w", err)
	}
	return &g, nil
}

func (r *Repository) UpdateGuest(ctx context.Context, g *domain.Guest) error {
	err := r.pool.QueryRow(ctx,
		`UPDATE guests SET
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
		 RETURNING id, tenant_id, first_name, last_name, email, phone,
		           address, city, state, country, pincode, nationality,
		           date_of_birth, aadhaar_number, passport_number, notes,
		           total_stays, created_at, updated_at`,
		g.ID, g.TenantID, g.FirstName, g.LastName, g.Email, g.Phone,
		g.Address, g.City, g.State, g.Country, g.Pincode, g.Nationality,
		g.DateOfBirth, g.AadhaarNumber, g.PassportNumber, g.Notes,
	).Scan(&g.ID, &g.TenantID, &g.FirstName, &g.LastName, &g.Email, &g.Phone,
		&g.Address, &g.City, &g.State, &g.Country, &g.Pincode, &g.Nationality,
		&g.DateOfBirth, &g.AadhaarNumber, &g.PassportNumber, &g.Notes,
		&g.TotalStays, &g.CreatedAt, &g.UpdatedAt)

	if err == pgx.ErrNoRows {
		return apperrors.NotFound("guest", g.ID.String())
	}
	if err != nil {
		return fmt.Errorf("update guest: %w", err)
	}
	return nil
}

func (r *Repository) ListGuests(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Guest, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM guests WHERE tenant_id = $1`, tenantID,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count guests: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, first_name, last_name, email, phone,
		        address, city, state, country, pincode, nationality,
		        date_of_birth, aadhaar_number, passport_number, notes,
		        total_stays, created_at, updated_at
		 FROM guests WHERE tenant_id = $1
		 ORDER BY last_name, first_name LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list guests: %w", err)
	}
	defer rows.Close()

	var guests []domain.Guest
	for rows.Next() {
		var g domain.Guest
		if err := rows.Scan(&g.ID, &g.TenantID, &g.FirstName, &g.LastName, &g.Email, &g.Phone,
			&g.Address, &g.City, &g.State, &g.Country, &g.Pincode, &g.Nationality,
			&g.DateOfBirth, &g.AadhaarNumber, &g.PassportNumber, &g.Notes,
			&g.TotalStays, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan guest: %w", err)
		}
		guests = append(guests, g)
	}

	if guests == nil {
		guests = []domain.Guest{}
	}

	return guests, count, nil
}

func (r *Repository) SearchGuests(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]domain.Guest, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM guests
		 WHERE tenant_id = $1
		   AND ($2::VARCHAR = '' OR first_name ILIKE '%' || $2 || '%'
		        OR last_name ILIKE '%' || $2 || '%'
		        OR phone ILIKE '%' || $2 || '%'
		        OR email ILIKE '%' || $2 || '%'
		        OR aadhaar_number = $2
		        OR passport_number = $2)`,
		tenantID, query,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count search guests: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, first_name, last_name, email, phone,
		        address, city, state, country, pincode, nationality,
		        date_of_birth, aadhaar_number, passport_number, notes,
		        total_stays, created_at, updated_at
		 FROM guests
		 WHERE tenant_id = $1
		   AND ($2::VARCHAR = '' OR first_name ILIKE '%' || $2 || '%'
		        OR last_name ILIKE '%' || $2 || '%'
		        OR phone ILIKE '%' || $2 || '%'
		        OR email ILIKE '%' || $2 || '%'
		        OR aadhaar_number = $2
		        OR passport_number = $2)
		 ORDER BY last_name, first_name LIMIT $3 OFFSET $4`,
		tenantID, query, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("search guests: %w", err)
	}
	defer rows.Close()

	var guests []domain.Guest
	for rows.Next() {
		var g domain.Guest
		if err := rows.Scan(&g.ID, &g.TenantID, &g.FirstName, &g.LastName, &g.Email, &g.Phone,
			&g.Address, &g.City, &g.State, &g.Country, &g.Pincode, &g.Nationality,
			&g.DateOfBirth, &g.AadhaarNumber, &g.PassportNumber, &g.Notes,
			&g.TotalStays, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan guest: %w", err)
		}
		guests = append(guests, g)
	}

	if guests == nil {
		guests = []domain.Guest{}
	}

	return guests, count, nil
}

func (r *Repository) IncrementGuestStays(ctx context.Context, guestID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE guests SET total_stays = total_stays + 1 WHERE id = $1`, guestID)
	return err
}

func (r *Repository) CreateGuestDocument(ctx context.Context, doc *domain.GuestDocument) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO guest_documents (tenant_id, guest_id, document_type, file_key, file_name, file_size, content_type)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, uploaded_at`,
		doc.TenantID, doc.GuestID, doc.DocumentType, doc.FileKey, doc.FileName, doc.FileSize, doc.ContentType,
	).Scan(&doc.ID, &doc.UploadedAt)

	if err != nil {
		return fmt.Errorf("create guest document: %w", err)
	}
	return nil
}

func (r *Repository) ListGuestDocuments(ctx context.Context, guestID, tenantID uuid.UUID) ([]domain.GuestDocument, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, guest_id, document_type, file_key, file_name, file_size, content_type, uploaded_at
		 FROM guest_documents WHERE guest_id = $1 AND tenant_id = $2 ORDER BY uploaded_at DESC`,
		guestID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list guest documents: %w", err)
	}
	defer rows.Close()

	var docs []domain.GuestDocument
	for rows.Next() {
		var d domain.GuestDocument
		if err := rows.Scan(&d.ID, &d.TenantID, &d.GuestID, &d.DocumentType, &d.FileKey,
			&d.FileName, &d.FileSize, &d.ContentType, &d.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, d)
	}

	if docs == nil {
		docs = []domain.GuestDocument{}
	}

	return docs, nil
}

func (r *Repository) GetGuestHistory(ctx context.Context, guestID, tenantID uuid.UUID, limit, offset int) ([]GuestStayRecord, int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM reservations WHERE guest_id = $1 AND tenant_id = $2`,
		guestID, tenantID,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count guest history: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT r.id, r.property_id, p.name, u.unit_number, ut.name,
		        r.check_in_date, r.check_out_date, r.status, r.total_amount, r.created_at
		 FROM reservations r
		 JOIN properties p ON r.property_id = p.id
		 JOIN units u ON r.unit_id = u.id
		 JOIN unit_types ut ON u.unit_type_id = ut.id
		 WHERE r.guest_id = $1 AND r.tenant_id = $2
		 ORDER BY r.check_in_date DESC LIMIT $3 OFFSET $4`,
		guestID, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("get guest history: %w", err)
	}
	defer rows.Close()

	var records []GuestStayRecord
	for rows.Next() {
		var rec GuestStayRecord
		if err := rows.Scan(&rec.ReservationID, &rec.PropertyID, &rec.PropertyName, &rec.UnitNumber, &rec.UnitTypeName,
			&rec.CheckInDate, &rec.CheckOutDate, &rec.Status, &rec.TotalAmount, &rec.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan guest history: %w", err)
		}
		records = append(records, rec)
	}

	if records == nil {
		records = []GuestStayRecord{}
	}

	return records, count, nil
}

type GuestStayRecord struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	PropertyID    uuid.UUID `json:"property_id"`
	PropertyName  string    `json:"property_name"`
	UnitNumber    string    `json:"unit_number"`
	UnitTypeName  string    `json:"unit_type_name"`
	CheckInDate   time.Time `json:"check_in_date"`
	CheckOutDate  time.Time `json:"check_out_date"`
	Status        string    `json:"status"`
	TotalAmount   float64   `json:"total_amount"`
	CreatedAt     time.Time `json:"created_at"`
}
