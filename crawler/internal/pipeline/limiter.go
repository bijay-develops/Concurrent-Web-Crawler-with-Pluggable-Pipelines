package pipeline

import (
	"sync"
	"time"
)

type DomainLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	limiters map[string]*time.Ticker
}

func NewDomainLimiter(interval time.Duration) *DomainLimiter {
	return &DomainLimiter{
		interval: interval,
		limiters: make(map[string]*time.Ticker),
	}
}

func (d *DomainLimiter) Wait(domain string) {
	d.mu.Lock()
	t, ok := d.limiters[domain]
	if !ok {
		t = time.NewTicker(d.interval)
		d.limiters[domain] = t
	}
	d.mu.Unlock()
	<-t.C
}
