package crawler

// WorkTracker is a simple utility to track ongoing work items in the crawler.
import "sync"

// WorkTracker is a simple wrapper around sync.WaitGroup to track the number of ongoing work items.
type WorkTracker struct {
	wg sync.WaitGroup
}

// NewWorkTracker creates a new WorkTracker instance.
func (w *WorkTracker) Add(n int) {
	w.wg.Add(n)
}

// Done signals that a work item is completed.
func (w *WorkTracker) Done() {
	w.wg.Done()
}

// Wait blocks until all work items have been completed.
func (w *WorkTracker) Wait() {
	w.wg.Wait()
}
