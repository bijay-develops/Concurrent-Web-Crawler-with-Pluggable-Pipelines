package pipeline

import "crawler/internal/shared"

type AllowAllFilter struct{}

func (AllowAllFilter) Allow(item shared.Item) bool {
	return true
}
