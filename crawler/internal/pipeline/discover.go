package pipeline

import (
	"context"
	"crawler/internal/shared"
	"net/url"
)

func DiscoverWorker(
	ctx context.Context,
	in <-chan shared.Item,
	out chan<- shared.Item,
	maxDepth int,
	tracker *shared.WorkTracker,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}

			// Schedule discovered internal links for further crawling.
			if item.Depth < maxDepth {
				for _, raw := range item.DiscoveredURLs {
					if raw == "" {
						continue
					}
					u, err := url.Parse(raw)
					if err != nil || u == nil {
						continue
					}
					// Work accounting: we add before enqueueing. The scheduler
					// will call Done() if it drops a duplicate.
					if tracker != nil {
						tracker.Add(1)
					}

					child := shared.Item{URL: u, Depth: item.Depth + 1, Mode: item.Mode}
					select {
					case out <- child:
						// ok
					case <-ctx.Done():
						// Undo bookkeeping for the child, and mark the current
						// item as done before exiting.
						if tracker != nil {
							tracker.Done()
							tracker.Done()
						}
						return
					}
				}
			}

			// Mark the current item as fully processed.
			if tracker != nil {
				tracker.Done()
			}
		}
	}
}
