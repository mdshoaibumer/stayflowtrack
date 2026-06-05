package middleware

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Metrics collects basic operational metrics in-memory.
// For production at scale, replace with Prometheus client.
type Metrics struct {
	TotalRequests    atomic.Int64
	TotalErrors      atomic.Int64
	ActiveRequests   atomic.Int64
	RequestDurations []time.Duration // circular buffer
}

var GlobalMetrics = &Metrics{}

// MetricsMiddleware tracks request counts and durations.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GlobalMetrics.TotalRequests.Add(1)
		GlobalMetrics.ActiveRequests.Add(1)
		defer GlobalMetrics.ActiveRequests.Add(-1)

		start := time.Now()
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		if ww.Status() >= 500 {
			GlobalMetrics.TotalErrors.Add(1)
		}

		_ = time.Since(start)
	})
}

// HealthHandler returns detailed health status including DB connectivity.
type HealthHandler struct {
	dbCheck func() error
}

func NewHealthHandler(dbCheck func() error) *HealthHandler {
	return &HealthHandler{dbCheck: dbCheck}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dbOK := "ok"
	if err := h.dbCheck(); err != nil {
		dbOK = "error: " + err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"` + dbOK + `","active_requests":` +
		strconv.FormatInt(GlobalMetrics.ActiveRequests.Load(), 10) +
		`,"total_requests":` + strconv.FormatInt(GlobalMetrics.TotalRequests.Load(), 10) +
		`,"total_errors":` + strconv.FormatInt(GlobalMetrics.TotalErrors.Load(), 10) + `}`))
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.dbCheck(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"ready":false}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ready":true}`))
}
