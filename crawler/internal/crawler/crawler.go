package crawler

import (
	"context"
	"errors"
)

type Config struct {
	WorkerCount int
}

type Crawler struct {
	cfg Config
}

func New(cfg Config) *Crawler {
	return &Crawler{
		cfg: cfg,
	}
}

func (c *Crawler) Run(ctx context.Context) error {
	if c.cfg.WorkerCount <= 0 {
		return errors.New("worker count must be > 0")
	}

	// Placeholder: real pipeline wiring comes next
	<-ctx.Done()

	return ctx.Err()
}
