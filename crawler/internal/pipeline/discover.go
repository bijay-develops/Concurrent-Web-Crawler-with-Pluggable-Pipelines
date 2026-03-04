package pipeline

import (
	"context"
	"crawler/internal/shared"
	"net/url"
	"strings"
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
				// Enqueue likely post links first for better results on
				// listing/tag pages.
				ordered := orderDiscoveredURLs(item.DiscoveredURLs)
				for _, raw := range ordered {
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

func orderDiscoveredURLs(urls []string) []string {
	if len(urls) == 0 {
		return nil
	}
	posts := make([]string, 0, len(urls))
	other := make([]string, 0, len(urls))
	for _, raw := range urls {
		p := strings.ToLower(raw)
		// Prefer typical post permalinks over listing/tag/category pages.
		if strings.Contains(p, "/tag/") || strings.Contains(p, "/category/") || strings.Contains(p, "/author/") {
			other = append(other, raw)
			continue
		}
		posts = append(posts, raw)
	}
	return append(posts, other...)
}
