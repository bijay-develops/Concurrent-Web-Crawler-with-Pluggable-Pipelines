package pipeline

import (
	"context"
	"crawler/internal/crawler"
)

func ParseWorker(
	ctx context.Context,
	in <-chan crawler.Item,
	out chan<- crawler.Item,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}
			if item.Response != nil {
				item.Response.Body.Close()
			}

			select {
			case out <- item:
			case <-ctx.Done():
				return 
			}
		}
	}
}
