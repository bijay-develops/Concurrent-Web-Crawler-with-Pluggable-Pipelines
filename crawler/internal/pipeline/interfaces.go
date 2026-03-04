package pipeline

import (
	"context"

	"crawler/internal/shared"
)

type Fetcher interface {
	Fetch(ctx context.Context, item shared.Item) (shared.Item, error)
}

type Parser interface {
	Parse(ctx context.Context, item shared.Item) ([]shared.Item, error)
}

type Filter interface {
	Allow(item shared.Item) bool
}

type Store interface {
	Store(ctx context.Context, item shared.Item) error
}
