package pipeline

import "crawler/internal/crawler"

type AllowAllFilter struct{}

func (AllowAllFilter) Allow(item crawler.Item) bool {
	return true
}
