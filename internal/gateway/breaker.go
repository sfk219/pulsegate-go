package gateway

import (
    "sync"
    "time"
)

// Breaker rejects traffic temporarily after consecutive upstream failures.
type Breaker struct {
    mu sync.Mutex
    consecutive int
    threshold int
    openUntil time.Time
    cooldown time.Duration
}

func NewBreaker(threshold int, cooldown time.Duration) *Breaker {
    if threshold < 1 { threshold = 3 }
    if cooldown <= 0 { cooldown = 10*time.Second }
    return &Breaker{threshold:threshold, cooldown:cooldown}
}

func (b *Breaker) Allow() bool {
    b.mu.Lock()
    defer b.mu.Unlock()
    if b.consecutive < b.threshold { return true }
    if time.Now().Before(b.openUntil) { return false }
    b.consecutive=0 // simple reset, not a complete half-open state machine
    return true
}
func (b *Breaker) Report(success bool) {
    b.mu.Lock()
    defer b.mu.Unlock()
    if success { b.consecutive=0; return }
    b.consecutive++
    if b.consecutive >= b.threshold { b.openUntil=time.Now().Add(b.cooldown) }
}
