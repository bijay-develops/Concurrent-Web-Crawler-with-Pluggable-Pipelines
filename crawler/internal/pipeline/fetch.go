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
	stats *shared.CrawlStats,
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

			// Send a more descriptive User-Agent so that some sites
			// that block generic bots have a better chance of accepting
			// the request.
			req.Header.Set("User-Agent", "ConcurrentWebCrawler/1.0 (+educational example)")

			resp, err := client.Do(req)
			if err != nil {
				if stats != nil {
					stats.RecordNetworkError()
				}
				log.Printf("[Crawl] request failed for %s: %v", item.URL.String(), err)
				continue
			}

			item.Response = resp
			if stats != nil {
				stats.RecordSuccess(item.URL.String(), resp.StatusCode)
			}

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
