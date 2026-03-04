package crawler

import (
	"context"
	"crawler/internal/shared"
	"sync"
)

// Scheduler is responsible for scheduling items to be processed by the crawler while ensuring that duplicate URLs are not processed multiple times.
type Scheduler struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

// NewScheduler creates a new Scheduler instance with an initialized seen map.
func NewScheduler() *Scheduler {
	return &Scheduler{
		seen: make(map[string]struct{}),
	}
}

// Schedule reads items from the input channel, checks for duplicates, and sends unique items to the output channel.
//
// If tracker is non-nil, it is used to keep accurate accounting of
// in-flight unique work items. Callers that already incremented the
// tracker for a candidate item can pass it here so duplicates are
// compensated via Done().
func (s *Scheduler) Schedule(ctx context.Context, in <-chan shared.Item, out chan<- shared.Item, tracker *shared.WorkTracker) {
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
				if tracker != nil {
					tracker.Done()
				}
				continue
			}
			s.seen[key] = struct{}{}
			s.mu.Unlock()

			// Send the unique item to the output channel, respecting the context cancellation.
			select {
			case out <- item:
			case <-ctx.Done():
				// Undo the bookkeeping if we were tracking this candidate.
				if tracker != nil {
					tracker.Done()
				}
				// Best-effort remove from seen so a future run could try again.
				s.mu.Lock()
				delete(s.seen, key)
				s.mu.Unlock()
				return
			}
		}
	}
}
