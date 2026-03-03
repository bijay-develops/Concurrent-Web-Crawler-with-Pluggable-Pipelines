package pipeline

import (
	"context"
	"log"

	"crawler/internal/crawler"
)

type LogStore struct{}

func (LogStore) Store(ctx context.Context, item crawler.Item) error {
	log.Println("stored:", item.URL.String())
	return nil
}
