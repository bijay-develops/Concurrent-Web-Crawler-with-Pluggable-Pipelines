package service

import (
	"context"
	"time"

	"crawler/internal/crawler"
	"crawler/internal/shared"
	"crawler/internal/store"
)

// StartRequest describes a crawl request coming from HTTP or CLI.
// This is a small DTO layer separating transport from the crawler package.
type StartRequest struct {
	URL     string
	Workers int
	Depth   int
	Mode    shared.UseCase
	Timeout time.Duration
}

// StartResponse is a simple summary of a finished crawl.
type StartResponse struct {
	URL     string
	Mode    shared.UseCase
	Stats   shared.CrawlStatsView
	Summary shared.ModeSummary
	Err     string
	Pages   []shared.PageRecord
}

// CrawlService coordinates running short-lived crawls for API callers.
// It optionally persists crawl summaries via a FileStore.
type CrawlService struct {
	store *store.FileStore
}

func NewCrawlService() *CrawlService {
	// Persist summaries under ./data/crawls.jsonl relative to the binary.
	return &CrawlService{store: store.NewFileStore("data/crawls.jsonl")}
}

// StartCrawl runs a crawl synchronously for now and returns aggregated stats.
// Later this can be extended to async jobs with IDs and persistent storage.
func (s *CrawlService) StartCrawl(ctx context.Context, req StartRequest) (StartResponse, error) {
	// Defaulting logic kept here so HTTP/CLI can stay thin.
	workers := req.Workers
	if workers <= 0 {
		workers = 8
	}

	depth := req.Depth
	if depth < 0 {
		depth = 2
	}

	mode := req.Mode
	if mode == "" {
		mode = shared.UseCaseTrackBlogs
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	stats := &shared.CrawlStats{}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c := crawler.New(
		crawler.WithWorkerCount(workers),
		crawler.WithMaxDepth(depth),
		crawler.WithSeedURL(req.URL),
		crawler.WithUseCase(mode),
		crawler.WithStatsCollector(stats),
	)

	var errStr string
	if err := c.Run(ctx); err != nil {
		// For the service we surface the error text but still return stats.
		// Callers can decide whether context cancellation is fatal.
		if err != context.Canceled && err != context.DeadlineExceeded {
			errStr = err.Error()
		}
	}

	view := stats.Snapshot()
	pages := stats.PagesSnapshot()

	rec := store.CrawlRecord{
		ID:         time.Now().UTC().Format("20060102T150405.000000000Z07:00"),
		StartedAt:  time.Now().UTC(),
		FinishedAt: time.Now().UTC(),
		URL:        req.URL,
		Mode:       mode,
		Stats:      view,
		Error:      errStr,
	}
	_ = s.store.SaveCrawl(ctx, rec)

	return StartResponse{
		URL:     req.URL,
		Mode:    mode,
		Stats:   view,
		Summary: shared.SummarizeMode(mode, view),
		Err:     errStr,
		Pages:   pages,
	}, nil
}
