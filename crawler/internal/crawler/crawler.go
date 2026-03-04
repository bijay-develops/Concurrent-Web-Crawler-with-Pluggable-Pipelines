package crawler

import (
	"context"
	"crawler/internal/pipeline"
	"crawler/internal/shared"
	"errors"
	"net/url"
	"sync"
	"time"
)

type Crawler struct {
	workers  int
	maxDepth int
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

func New(opts ...Option) *Crawler {
	c := &Crawler{
		workers:  4,
		maxDepth: 1,
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
			pipeline.FetchWorker(ctx, client, limiter, scheduled, fetched)
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
		u, _ := url.Parse("https://example.com")
		tracker.Add(1)
		seeds <- shared.Item{URL: u, Depth: 0}
	}()

	<-ctx.Done()
	return ctx.Err()
}
