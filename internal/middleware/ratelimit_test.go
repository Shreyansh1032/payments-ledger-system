package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/Shreyansh1032/payments-ledger-system/internal/middleware"
)

func TestRateLimiter_BlocksAfterBurst(t *testing.T) {
	limiter := middleware.NewRateLimiter(rate.Every(time.Minute), 3)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.RemoteAddr = "192.0.2.1:12345"

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 on 4th request, got %d", rec.Code)
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	limiter := middleware.NewRateLimiter(rate.Every(time.Minute), 1)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req1.RemoteAddr = "192.0.2.1:12345"
	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req2.RemoteAddr = "192.0.2.2:12345"

	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("ip1 first request: expected 200, got %d", rec1.Code)
	}

	rec1b := httptest.NewRecorder()
	handler.ServeHTTP(rec1b, req1)
	if rec1b.Code != http.StatusTooManyRequests {
		t.Fatalf("ip1 second request: expected 429, got %d", rec1b.Code)
	}

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("ip2 first request: expected 200 (independent from ip1's limit), got %d", rec2.Code)
	}
}
