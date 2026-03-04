package pipeline

import (
	"context"
	"crawler/internal/shared"
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

			tracker.Done()

			if item.Depth >= maxDepth {
				continue
			}

			// placeholder: no real discovery yet
			_ = out
		}
	}
}
