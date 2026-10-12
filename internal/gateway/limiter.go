package gateway

import (
    "math"
    "sync"
    "time"
)

type bucket struct { tokens float64; updated time.Time }

// Limiter is an in-memory per-client token bucket. Keys are not garbage-collected.
type Limiter struct {
    mu sync.Mutex
    buckets map[string]*bucket
    rate float64
    burst float64
}

func NewLimiter(rate float64, burst int) *Limiter {
    if rate <= 0 { rate = 5 }
    if burst < 1 { burst = 10 }
    return &Limiter{buckets: make(map[string]*bucket), rate:rate, burst:float64(burst)}
}

func (l *Limiter) Allow(key string) bool {
    now := time.Now()
    l.mu.Lock()
    defer l.mu.Unlock()
    b, ok := l.buckets[key]
    if !ok { l.buckets[key] = &bucket{tokens:l.burst-1, updated:now}; return true }
    b.tokens = math.Min(l.burst, b.tokens + now.Sub(b.updated).Seconds()*l.rate)
    b.updated = now
    if b.tokens < 1 { return false }
    b.tokens--
    return true
}
