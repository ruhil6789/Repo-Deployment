package ratelimit

import (
	"sync"
	"time"
)

// Limiter implements a simple token bucket rate limiter
type Limiter struct {
	rate       int           // requests per duration
	duration   time.Duration // time window
	tokens     int           // current tokens
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewLimiter creates a new rate limiter
func NewLimiter(rate int, duration time.Duration) *Limiter {
	return &Limiter{
		rate:       rate,
		duration:   duration,
		tokens:     rate,
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request is allowed
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastUpdate)

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed / (l.duration / time.Duration(l.rate)))
	if tokensToAdd > 0 {
		l.tokens = min(l.tokens+tokensToAdd, l.rate)
		l.lastUpdate = now
	}

	if l.tokens > 0 {
		l.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
