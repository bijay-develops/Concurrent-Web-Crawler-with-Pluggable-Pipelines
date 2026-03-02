package pipeline

import (
	"context"
	"crawler/internal/crawler"
)

func DiscoverWorker(
	ctx context.Context,
	in <-chan crawler.Item,
	out chan<- crawler.Item,
	maxDepth int,
	tracker *crawler.WorkTracker,
){
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}

			tracker.Done()

			if item.Depth >= maxDepth {
				continue
			}

			// placeholder: no real discovery yet
			_= out
		}
	}
}