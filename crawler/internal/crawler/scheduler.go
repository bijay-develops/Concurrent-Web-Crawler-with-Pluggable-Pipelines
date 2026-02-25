package crawler

import (
	"context"
	"sync"
)

// Schedular is responsible for scheduling items to be processed by the crawler while ensuring that duplicate URLs are not processed multiple times.
type Schedular struct {
	mu sync.Mutex
	seen map[string]struct{}
}
// NewSchedular creates a new Schedular instance with an initialized seen map.
func NewSchedular() *Schedular {
	return &Schedular{
		seen: make(map[string]struct{}),
	}
}

// Schedule reads items from the input channel, checks for duplicates, and sends unique items to the output channel.
func (s *Schedular) Schedule(ctx context.Context, in <-chan Item, out chan<- Item) {
	for {
// Use a select statement to handle context cancellation and channel input simultaneously.
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}

// Use the URL string as the key to check for duplicates.
			key := item.URL.String()

// Lock the mutex to safely access the seen map.
			s.mu.Lock()
			if _, exists := s.seen[key]; exists {
				s.mu.Unlock()
				continue
			}
			s.seen[key] = struct{}{}
			s.mu.Unlock()

// Send the unique item to the output channel, respecting the context cancellation.
			select {
			case out <- item:
			case <-ctx.Done():
				return
			}
		}
	}
}
