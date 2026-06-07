package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    3,
		window:   1 * time.Minute,
	}

	for i := 0; i < 3; i++ {
		if !rl.isAllowed("192.168.1.1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    2,
		window:   1 * time.Minute,
	}

	rl.isAllowed("10.0.0.1")
	rl.isAllowed("10.0.0.1")

	if rl.isAllowed("10.0.0.1") {
		t.Error("third request should be blocked")
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    1,
		window:   1 * time.Minute,
	}

	if !rl.isAllowed("10.0.0.1") {
		t.Error("first IP first request should be allowed")
	}
	if !rl.isAllowed("10.0.0.2") {
		t.Error("second IP first request should be allowed")
	}
	if rl.isAllowed("10.0.0.1") {
		t.Error("first IP second request should be blocked")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    1,
		window:   50 * time.Millisecond,
	}

	rl.isAllowed("10.0.0.1")
	if rl.isAllowed("10.0.0.1") {
		t.Error("should be blocked immediately")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.isAllowed("10.0.0.1") {
		t.Error("should be allowed after window expires")
	}
}

func TestRateLimiter_LimitMiddleware_Returns429(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    1,
		window:   1 * time.Minute,
	}

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request succeeds
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rr.Code)
	}

	// Second request rate limited
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") != "60" {
		t.Error("expected Retry-After header")
	}
}

func TestRateLimiter_UsesXForwardedFor(t *testing.T) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    1,
		window:   1 * time.Minute,
	}

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("second request should be rate limited by X-Forwarded-For IP")
	}
}

func TestMaxBodySize_AllowsSmallBody(t *testing.T) {
	handler := MaxBodySize(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("small body"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMaxBodySize_RejectsLargeBody(t *testing.T) {
	handler := MaxBodySize(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	largeBody := strings.Repeat("x", 100)
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(largeBody))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", rr.Code)
	}
}

func TestMetricsMiddleware_TracksRequests(t *testing.T) {
	// Reset global metrics
	GlobalMetrics.TotalRequests.Store(0)
	GlobalMetrics.TotalErrors.Store(0)
	GlobalMetrics.ActiveRequests.Store(0)
	GlobalMetrics.status2xx.Store(0)
	GlobalMetrics.status5xx.Store(0)

	handler := MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if GlobalMetrics.TotalRequests.Load() != 1 {
		t.Errorf("expected TotalRequests=1, got %d", GlobalMetrics.TotalRequests.Load())
	}
	if GlobalMetrics.status2xx.Load() != 1 {
		t.Errorf("expected status2xx=1, got %d", GlobalMetrics.status2xx.Load())
	}
	if GlobalMetrics.ActiveRequests.Load() != 0 {
		t.Error("expected ActiveRequests=0 after request completes")
	}
}

func TestMetricsMiddleware_Tracks5xx(t *testing.T) {
	GlobalMetrics.TotalErrors.Store(0)
	GlobalMetrics.status5xx.Store(0)

	handler := MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if GlobalMetrics.TotalErrors.Load() != 1 {
		t.Errorf("expected TotalErrors=1, got %d", GlobalMetrics.TotalErrors.Load())
	}
	if GlobalMetrics.status5xx.Load() != 1 {
		t.Errorf("expected status5xx=1, got %d", GlobalMetrics.status5xx.Load())
	}
}

func TestMetricsHandler_ReturnsPrometheusFormat(t *testing.T) {
	GlobalMetrics.TotalRequests.Store(42)
	GlobalMetrics.TotalErrors.Store(2)
	GlobalMetrics.ActiveRequests.Store(0)
	GlobalMetrics.status2xx.Store(35)
	GlobalMetrics.status4xx.Store(5)
	GlobalMetrics.status5xx.Store(2)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	MetricsHandler(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "http_requests_total 42") {
		t.Error("missing http_requests_total metric")
	}
	if !strings.Contains(body, "http_errors_total 2") {
		t.Error("missing http_errors_total metric")
	}
	if !strings.Contains(body, `http_responses_total{code="2xx"} 35`) {
		t.Error("missing 2xx response metric")
	}
	if !strings.Contains(body, "http_request_duration_seconds_bucket") {
		t.Error("missing histogram bucket metrics")
	}
	if rr.Header().Get("Content-Type") != "text/plain; version=0.0.4; charset=utf-8" {
		t.Errorf("unexpected content-type: %s", rr.Header().Get("Content-Type"))
	}
}

func TestHealthHandler_Healthy(t *testing.T) {
	h := NewHealthHandler(func() error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
		t.Error("expected status ok in response")
	}
}

func TestHealthHandler_Unhealthy(t *testing.T) {
	h := NewHealthHandler(func() error { return io.ErrUnexpectedEOF })

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.Health(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestReadyHandler_Ready(t *testing.T) {
	h := NewHealthHandler(func() error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()
	h.Ready(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"ready":true`) {
		t.Error("expected ready:true")
	}
}

func TestReadyHandler_NotReady(t *testing.T) {
	h := NewHealthHandler(func() error { return io.ErrUnexpectedEOF })

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()
	h.Ready(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestFilterRecent(t *testing.T) {
	now := time.Now()
	window := 1 * time.Minute

	timestamps := []time.Time{
		now.Add(-2 * time.Minute),  // expired
		now.Add(-30 * time.Second), // recent
		now.Add(-10 * time.Second), // recent
	}

	result := filterRecent(timestamps, now, window)
	if len(result) != 2 {
		t.Errorf("expected 2 recent timestamps, got %d", len(result))
	}
}
