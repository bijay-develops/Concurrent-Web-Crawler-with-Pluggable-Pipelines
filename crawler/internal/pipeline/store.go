package pipeline

import (
	"context"
	"log"

	"crawler/internal/shared"
)

type LogStore struct{}

func (LogStore) Store(ctx context.Context, item shared.Item) error {
	log.Println("stored:", item.URL.String())
	return nil
}
