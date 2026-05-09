package crawler

import (
	"context"
	"sync"
	"time"
)

type requestLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
}

func newRequestLimiter(rate float64) *requestLimiter {
	if rate <= 0 {
		return nil
	}

	interval := time.Duration(float64(time.Second) / rate)
	if interval <= 0 {
		interval = time.Nanosecond
	}

	return &requestLimiter{interval: interval}
}

func (l *requestLimiter) wait(ctx context.Context) error {
	if l == nil {
		return nil
	}

	wait := l.reserveDelay(time.Now())
	if wait <= 0 {
		return nil
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (l *requestLimiter) reserveDelay(now time.Time) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.next.IsZero() || now.After(l.next) {
		l.next = now.Add(l.interval)

		return 0
	}

	wait := l.next.Sub(now)
	l.next = l.next.Add(l.interval)

	return wait
}

func (l *requestLimiter) stop() {}
