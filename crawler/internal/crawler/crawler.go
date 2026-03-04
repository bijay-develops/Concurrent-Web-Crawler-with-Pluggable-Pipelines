package crawler

import (
	"context"
	"crawler/internal/pipeline"
	"crawler/internal/shared"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"
)

type Crawler struct {
	workers  int
	maxDepth int
	seedURL  string
	mode     shared.UseCase
}

type Option func(*Crawler)

func WithWorkerCount(n int) Option {
	return func(c *Crawler) {
		c.workers = n
	}
}

func WithMaxDepth(d int) Option {
	return func(c *Crawler) {
		c.maxDepth = d
	}
}

func WithSeedURL(u string) Option {
	return func(c *Crawler) {
		c.seedURL = u
	}
}

func WithUseCase(mode shared.UseCase) Option {
	return func(c *Crawler) {
		c.mode = mode
	}
}

func New(opts ...Option) *Crawler {
	c := &Crawler{
		workers:  4,
		maxDepth: 1,
		seedURL:  "https://example.com",
		mode:     shared.UseCaseTrackBlogs,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Crawler) Run(ctx context.Context) error {
	if c.workers <= 0 {
		return errors.New("worker count must be > 0")
	}

	seedURL, err := parseSeedURL(c.seedURL)
	if err != nil {
		return err
	}

	tracker := &shared.WorkTracker{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	seeds := make(chan shared.Item)
	scheduled := make(chan shared.Item)
	fetched := make(chan shared.Item)
	parsed := make(chan shared.Item)
	discovered := make(chan shared.Item)

	scheduler := NewScheduler()

	go scheduler.Schedule(ctx, seeds, scheduled)

	client := pipeline.NewHTTPClient(10 * time.Second)
	limiter := pipeline.NewDomainLimiter(500 * time.Millisecond)

	var fetchWG sync.WaitGroup
	fetchWG.Add(c.workers)
	for i := 0; i < c.workers; i++ {
		go func() {
			defer fetchWG.Done()
			pipeline.FetchWorker(ctx, client, limiter, scheduled, fetched, c.mode)
		}()
	}

	go func() {
		fetchWG.Wait()
		close(fetched)
	}()

	go pipeline.ParseWorker(ctx, fetched, parsed)
	go pipeline.DiscoverWorker(ctx, parsed, discovered, c.maxDepth, tracker)
	go scheduler.Schedule(ctx, discovered, scheduled)

	go func() {
		tracker.Wait()
		cancel()
	}()

	go func() {
		defer close(seeds)
		tracker.Add(1)
		seeds <- shared.Item{URL: seedURL, Depth: 0, Mode: c.mode}
	}()

	<-ctx.Done()
	return ctx.Err()
}

func parseSeedURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid seed URL %q: %w", raw, err)
	}

	// If the user passed something like "facebook.com", assume https.
	if parsed.Scheme == "" {
		parsed, err = url.Parse("https://" + raw)
		if err != nil {
			return nil, fmt.Errorf("invalid seed URL %q after adding https://: %w", raw, err)
		}
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("seed URL %q must include a host", raw)
	}

	return parsed, nil
}
