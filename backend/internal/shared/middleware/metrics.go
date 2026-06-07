package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Metrics collects operational metrics exposed via /metrics in Prometheus text format.
type Metrics struct {
	TotalRequests  atomic.Int64
	TotalErrors    atomic.Int64
	ActiveRequests atomic.Int64

	// Histogram buckets for request duration (seconds)
	durationBuckets  [7]atomic.Int64 // <=5ms, <=25ms, <=100ms, <=250ms, <=1s, <=5s, >5s
	durationSumMicro atomic.Int64    // sum in microseconds
	durationCount    atomic.Int64

	// Status code counters
	status2xx atomic.Int64
	status3xx atomic.Int64
	status4xx atomic.Int64
	status5xx atomic.Int64
}

var GlobalMetrics = &Metrics{}

var bucketBounds = [6]time.Duration{
	5 * time.Millisecond,
	25 * time.Millisecond,
	100 * time.Millisecond,
	250 * time.Millisecond,
	1 * time.Second,
	5 * time.Second,
}

// MetricsMiddleware tracks request counts, durations, and status codes.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GlobalMetrics.TotalRequests.Add(1)
		GlobalMetrics.ActiveRequests.Add(1)
		defer GlobalMetrics.ActiveRequests.Add(-1)

		start := time.Now()
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		GlobalMetrics.durationSumMicro.Add(duration.Microseconds())
		GlobalMetrics.durationCount.Add(1)

		// Bucket assignment
		switch {
		case duration <= bucketBounds[0]:
			GlobalMetrics.durationBuckets[0].Add(1)
		case duration <= bucketBounds[1]:
			GlobalMetrics.durationBuckets[1].Add(1)
		case duration <= bucketBounds[2]:
			GlobalMetrics.durationBuckets[2].Add(1)
		case duration <= bucketBounds[3]:
			GlobalMetrics.durationBuckets[3].Add(1)
		case duration <= bucketBounds[4]:
			GlobalMetrics.durationBuckets[4].Add(1)
		case duration <= bucketBounds[5]:
			GlobalMetrics.durationBuckets[5].Add(1)
		default:
			GlobalMetrics.durationBuckets[6].Add(1)
		}

		status := ww.Status()
		switch {
		case status >= 500:
			GlobalMetrics.TotalErrors.Add(1)
			GlobalMetrics.status5xx.Add(1)
		case status >= 400:
			GlobalMetrics.status4xx.Add(1)
		case status >= 300:
			GlobalMetrics.status3xx.Add(1)
		default:
			GlobalMetrics.status2xx.Add(1)
		}
	})
}

// MetricsHandler exposes metrics in Prometheus text exposition format.
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	m := GlobalMetrics
	total := m.TotalRequests.Load()
	errors := m.TotalErrors.Load()
	active := m.ActiveRequests.Load()
	count := m.durationCount.Load()
	sumSec := float64(m.durationSumMicro.Load()) / 1e6

	// Cumulative bucket counts for histogram
	var cum int64
	var buckets [7]int64
	for i := range m.durationBuckets {
		cum += m.durationBuckets[i].Load()
		buckets[i] = cum
	}

	_, _ = fmt.Fprintf(w, "# HELP http_requests_total Total number of HTTP requests.\n")
	_, _ = fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
	_, _ = fmt.Fprintf(w, "http_requests_total %d\n", total)

	_, _ = fmt.Fprintf(w, "# HELP http_requests_active Current in-flight requests.\n")
	_, _ = fmt.Fprintf(w, "# TYPE http_requests_active gauge\n")
	_, _ = fmt.Fprintf(w, "http_requests_active %d\n", active)

	_, _ = fmt.Fprintf(w, "# HELP http_errors_total Total 5xx responses.\n")
	_, _ = fmt.Fprintf(w, "# TYPE http_errors_total counter\n")
	_, _ = fmt.Fprintf(w, "http_errors_total %d\n", errors)

	_, _ = fmt.Fprintf(w, "# HELP http_responses_total Responses by status class.\n")
	_, _ = fmt.Fprintf(w, "# TYPE http_responses_total counter\n")
	_, _ = fmt.Fprintf(w, "http_responses_total{code=\"2xx\"} %d\n", m.status2xx.Load())
	_, _ = fmt.Fprintf(w, "http_responses_total{code=\"3xx\"} %d\n", m.status3xx.Load())
	_, _ = fmt.Fprintf(w, "http_responses_total{code=\"4xx\"} %d\n", m.status4xx.Load())
	_, _ = fmt.Fprintf(w, "http_responses_total{code=\"5xx\"} %d\n", m.status5xx.Load())

	_, _ = fmt.Fprintf(w, "# HELP http_request_duration_seconds Request duration histogram.\n")
	_, _ = fmt.Fprintf(w, "# TYPE http_request_duration_seconds histogram\n")
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=\"0.005\"} %d\n", buckets[0])
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=\"0.025\"} %d\n", buckets[1])
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=\"0.1\"} %d\n", buckets[2])
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=\"0.25\"} %d\n", buckets[3])
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=\"1\"} %d\n", buckets[4])
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=\"5\"} %d\n", buckets[5])
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=\"+Inf\"} %d\n", buckets[6])
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_sum %.6f\n", sumSec)
	_, _ = fmt.Fprintf(w, "http_request_duration_seconds_count %d\n", count)
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
