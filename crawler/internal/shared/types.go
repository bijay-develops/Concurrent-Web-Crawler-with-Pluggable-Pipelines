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

// CrawlStats holds simple, user-friendly statistics about a crawl.
// It is safe for concurrent use by multiple workers.
type CrawlStats struct {
	mu sync.Mutex

	TotalRequests  int
	Success2xx     int
	ClientError4xx int
	ServerError5xx int
	OtherStatus    int
	NetworkErrors  int

	LastStatusCode int
	LastURL        string
}

// RecordSuccess updates stats for a successful HTTP response.
func (s *CrawlStats) RecordSuccess(url string, status int) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalRequests++
	s.LastStatusCode = status
	s.LastURL = url

	switch {
	case status >= 200 && status < 300:
		s.Success2xx++
	case status >= 400 && status < 500:
		s.ClientError4xx++
	case status >= 500 && status < 600:
		s.ServerError5xx++
	default:
		s.OtherStatus++
	}
}

// RecordNetworkError updates stats when a request fails before getting a response.
func (s *CrawlStats) RecordNetworkError() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalRequests++
	s.NetworkErrors++
}

// Snapshot returns a copy of the stats for read-only use.
func (s *CrawlStats) Snapshot() CrawlStats {
	if s == nil {
		return CrawlStats{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return CrawlStats{
		TotalRequests:  s.TotalRequests,
		Success2xx:     s.Success2xx,
		ClientError4xx: s.ClientError4xx,
		ServerError5xx: s.ServerError5xx,
		OtherStatus:    s.OtherStatus,
		NetworkErrors:  s.NetworkErrors,
		LastStatusCode: s.LastStatusCode,
		LastURL:        s.LastURL,
	}
}
