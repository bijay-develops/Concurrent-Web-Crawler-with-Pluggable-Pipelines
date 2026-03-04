package pipeline

import (
	"context"
	"log"
	"net/http"
	"time"

	"crawler/internal/shared"
)

func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func FetchWorker(
	ctx context.Context,
	client *http.Client,
	limiter *DomainLimiter,
	in <-chan shared.Item,
	out chan<- shared.Item,
	mode shared.UseCase,
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

			switch mode {
			case shared.UseCaseTrackBlogs:
				log.Printf("[Blogs] fetched %s status=%d", item.URL.String(), resp.StatusCode)
			case shared.UseCaseSiteHealth:
				log.Printf("[Health] fetched %s status=%d", item.URL.String(), resp.StatusCode)
			case shared.UseCaseSearchIndex:
				log.Printf("[SearchIndex] fetched %s status=%d", item.URL.String(), resp.StatusCode)
			default:
				log.Printf("[Crawl] fetched %s status=%d", item.URL.String(), resp.StatusCode)
			}

			select {
			case out <- item:
			case <-ctx.Done():
				resp.Body.Close()
				return
			}
		}
	}
}
