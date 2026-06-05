package audit

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Action constants for audit logging.
const (
	ActionCreate        = "create"
	ActionUpdate        = "update"
	ActionDelete        = "delete"
	ActionStatusChange  = "status_change"
	ActionLogin         = "login"
	ActionLogout        = "logout"
	ActionPasswordReset = "password_reset"
	ActionCheckIn       = "check_in"
	ActionCheckOut      = "check_out"
	ActionPayment       = "payment"
	ActionRefund        = "refund"
	ActionVoid          = "void"
)

// Entry represents an audit log entry.
type Entry struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	OldValues  map[string]any
	NewValues  map[string]any
	IPAddress  string
	UserAgent  string
}

// Logger records audit entries to the database.
type Logger struct {
	pool *pgxpool.Pool
}

// New creates an audit logger.
func New(pool *pgxpool.Pool) *Logger {
	return &Logger{pool: pool}
}

// Log records an audit entry asynchronously (fire and forget for non-critical path).
func (l *Logger) Log(ctx context.Context, entry Entry) {
	go l.write(context.Background(), entry)
}

// LogSync records an audit entry synchronously (for critical operations).
func (l *Logger) LogSync(ctx context.Context, entry Entry) error {
	return l.write(ctx, entry)
}

func (l *Logger) write(ctx context.Context, entry Entry) error {
	_, err := l.pool.Exec(ctx,
		`INSERT INTO audit_logs (tenant_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		entry.TenantID, nilUUID(entry.UserID), entry.Action, entry.EntityType,
		nilUUID(entry.EntityID), jsonOrNil(entry.OldValues), jsonOrNil(entry.NewValues),
		nilString(entry.IPAddress), nilString(entry.UserAgent),
	)
	return err
}

// FromRequest extracts IP and User-Agent from an HTTP request.
func FromRequest(r *http.Request) (ip string, userAgent string) {
	ip = r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = fwd
	}
	return ip, r.UserAgent()
}

func nilUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func nilString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func jsonOrNil(m map[string]any) any {
	if len(m) == 0 {
		return nil
	}
	return m
}
