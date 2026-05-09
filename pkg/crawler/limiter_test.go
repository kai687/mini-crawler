package crawler

import (
	"testing"
	"time"
)

func TestRequestLimiterAllowsFirstRequestImmediately(t *testing.T) {
	limiter := newRequestLimiter(10)
	now := time.Unix(100, 0)

	if got := limiter.reserveDelay(now); got != 0 {
		t.Fatalf("reserveDelay() = %s, want 0", got)
	}
}

func TestRequestLimiterSpacesRequests(t *testing.T) {
	limiter := newRequestLimiter(10)
	now := time.Unix(100, 0)

	_ = limiter.reserveDelay(now)

	if got := limiter.reserveDelay(now); got != 100*time.Millisecond {
		t.Fatalf("second reserveDelay() = %s, want 100ms", got)
	}

	if got := limiter.reserveDelay(now); got != 200*time.Millisecond {
		t.Fatalf("third reserveDelay() = %s, want 200ms", got)
	}
}

func TestRequestLimiterDisabledForZeroRate(t *testing.T) {
	if got := newRequestLimiter(0); got != nil {
		t.Fatalf("newRequestLimiter(0) = %#v, want nil", got)
	}
}
