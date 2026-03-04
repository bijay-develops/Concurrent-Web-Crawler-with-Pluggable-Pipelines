package pipeline

import (
	"sync"
	"time"
)

type DomainLimiter struct {
	mu       sync.Mutex
	interval time.Duration

	// lastScheduled stores the last scheduled time for a given domain.
	// Each Wait() call reserves the next slot so concurrent goroutines
	// naturally serialize per-domain without spawning tickers.
	lastScheduled map[string]time.Time

	// calls is used to trigger periodic cleanup.
	calls int
}

func NewDomainLimiter(interval time.Duration) *DomainLimiter {
	return &DomainLimiter{
		interval:      interval,
		lastScheduled: make(map[string]time.Time),
	}
}

func (d *DomainLimiter) Wait(domain string) {
	if d == nil {
		return
	}
	if d.interval <= 0 {
		return
	}
	if domain == "" {
		// No domain; nothing to rate-limit.
		return
	}

	now := time.Now()

	d.mu.Lock()
	last := d.lastScheduled[domain]
	scheduled := now
	if !last.IsZero() {
		next := last.Add(d.interval)
		if next.After(now) {
			scheduled = next
		}
	}
	d.lastScheduled[domain] = scheduled

	// Periodic cleanup to avoid unbounded growth if many domains are seen.
	d.calls++
	if d.calls%500 == 0 {
		cutoff := now.Add(-10 * d.interval)
		for k, t := range d.lastScheduled {
			if t.Before(cutoff) {
				delete(d.lastScheduled, k)
			}
		}
	}
	d.mu.Unlock()

	if delay := time.Until(scheduled); delay > 0 {
		time.Sleep(delay)
	}
}
