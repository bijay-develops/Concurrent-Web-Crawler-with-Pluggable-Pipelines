package pipeline

import (
	"context"
	"net/http"
	"time"

	"crawler/internal/crawler"
)

func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func FetchWorker(
	ctx context.Context,
	client *http.Client,
	limiter *DomainLimiter,
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

			limiter.Wait(item.URL.Host)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.URL.String(), nil)
			if err != nil {
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				continue
			}

			item.Response = resp

			select {
			case out <- item:
			case <-ctx.Done():
				resp.Body.Close()
				return
			}
		}
	}
}
