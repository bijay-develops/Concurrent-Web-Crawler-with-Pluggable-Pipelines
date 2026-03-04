package shared

import (
	"net/http"
	"net/url"
	"sync"
)

// UseCase describes the high-level task the crawler is performing.
// It allows the rest of the system (workers, logging, storage) to adapt
// behavior without changing core orchestration logic.
type UseCase string

const (
	UseCaseTrackBlogs  UseCase = "track-blogs"
	UseCaseSiteHealth  UseCase = "site-health"
	UseCaseSearchIndex UseCase = "search-index"
)

type Item struct {
	URL      *url.URL
	Depth    int
	Response *http.Response
	Mode     UseCase
}

type WorkTracker struct {
	wg sync.WaitGroup
}

func (w *WorkTracker) Add(n int) {
	w.wg.Add(n)
}

func (w *WorkTracker) Done() {
	w.wg.Done()
}

func (w *WorkTracker) Wait() {
	w.wg.Wait()
}
