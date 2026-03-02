package pipeline

import (
	"context"

	"crawler/internal/crawler"
)

type Fetcher interface {
	Fetch(ctx context.Context, item crawler.Item) (crawler.Item, error)
}

type Parser interface {
	Parse(ctx context.Context, item crawler.Item) ([]crawler.Item, error)
}

type Filter interface {
	Allow(item crawler.Item) bool
}

type Store interface {
	Store(ctx context.Context, item crawler.Item) error
}
