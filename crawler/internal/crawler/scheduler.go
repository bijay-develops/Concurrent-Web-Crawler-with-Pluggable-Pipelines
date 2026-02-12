package crawler

import (
	"context"
	"sync"
)

type Schedular struct {
	seen map[string]struct{}
	mu   sync.Mutex
}

func NewSchedular() *Schedular {
	return &Schedular{
		seen: make(map[string]struct{}),
	}
}

func (s *Schedular) Seen(u string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.seen[u]; ok {
		return true
	}

	s.seen[u] = struct{}{}
	return false
}

func (s *Schedular) Run(ctx context.Context, in <-chan Item, out chan<- Item) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}
			if s.Seen(item.URL.String()) {
				continue
			}
			out <- item
		}
	}
}
