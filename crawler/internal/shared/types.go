package shared

import (
	"fmt"
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

	// Simple scraping / content analytics
	ParsedPages          int
	TotalWords           int
	TotalInternalLinks   int
	TotalExternalLinks   int
	LongestPageWordCount int
	LongestPageURL       string
	LongestPageTitle     string
}

// CrawlStatsView is an immutable, lock-free snapshot of CrawlStats
// intended for read-only reporting and UI rendering.
type CrawlStatsView struct {
	TotalRequests  int `json:"totalRequests"`
	Success2xx     int `json:"success2xx"`
	ClientError4xx int `json:"clientError4xx"`
	ServerError5xx int `json:"serverError5xx"`
	OtherStatus    int `json:"otherStatus"`
	NetworkErrors  int `json:"networkErrors"`

	LastStatusCode int    `json:"lastStatusCode"`
	LastURL        string `json:"lastUrl"`

	ParsedPages          int    `json:"parsedPages"`
	TotalWords           int    `json:"totalWords"`
	TotalInternalLinks   int    `json:"totalInternalLinks"`
	TotalExternalLinks   int    `json:"totalExternalLinks"`
	LongestPageWordCount int    `json:"longestPageWordCount"`
	LongestPageURL       string `json:"longestPageUrl"`
	LongestPageTitle     string `json:"longestPageTitle"`
}

// ModeSummary is a small, user-friendly interpretation of stats for a
// particular high-level use case. It is designed to be sent over APIs
// and rendered directly in UIs.
type ModeSummary struct {
	Mode          UseCase        `json:"mode"`
	CheckedPages  int            `json:"checkedPages"`
	IsReachable   bool           `json:"isReachable"`
	IsHealthy     bool           `json:"isHealthy"`
	IsIndexable   bool           `json:"isIndexable"`
	PrimaryStatus int            `json:"primaryStatus"`
	Message       string         `json:"message"`
	RawStats      CrawlStatsView `json:"rawStats"`

	// Extra analytics derived from RawStats for convenience in UIs.
	AverageWordsPerPage  int    `json:"averageWordsPerPage"`
	LongestPageURL       string `json:"longestPageUrl"`
	LongestPageTitle     string `json:"longestPageTitle"`
	LongestPageWordCount int    `json:"longestPageWordCount"`
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

// RecordPageMetrics updates simple scraping/analytics statistics for a
// successfully fetched page. It is safe for concurrent use.
func (s *CrawlStats) RecordPageMetrics(url, title string, wordCount, internalLinks, externalLinks int) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if wordCount > 0 {
		s.ParsedPages++
		s.TotalWords += wordCount
		if wordCount > s.LongestPageWordCount {
			s.LongestPageWordCount = wordCount
			s.LongestPageURL = url
			s.LongestPageTitle = title
		}
	}

	if internalLinks > 0 {
		s.TotalInternalLinks += internalLinks
	}
	if externalLinks > 0 {
		s.TotalExternalLinks += externalLinks
	}
}

// Snapshot returns a lock-free copy of the stats for read-only use.
func (s *CrawlStats) Snapshot() CrawlStatsView {
	if s == nil {
		return CrawlStatsView{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return CrawlStatsView{
		TotalRequests:        s.TotalRequests,
		Success2xx:           s.Success2xx,
		ClientError4xx:       s.ClientError4xx,
		ServerError5xx:       s.ServerError5xx,
		OtherStatus:          s.OtherStatus,
		NetworkErrors:        s.NetworkErrors,
		LastStatusCode:       s.LastStatusCode,
		LastURL:              s.LastURL,
		ParsedPages:          s.ParsedPages,
		TotalWords:           s.TotalWords,
		TotalInternalLinks:   s.TotalInternalLinks,
		TotalExternalLinks:   s.TotalExternalLinks,
		LongestPageWordCount: s.LongestPageWordCount,
		LongestPageURL:       s.LongestPageURL,
		LongestPageTitle:     s.LongestPageTitle,
	}
}

// SummarizeMode converts raw stats into a simple, human-readable
// interpretation for the given use case.
func SummarizeMode(mode UseCase, v CrawlStatsView) ModeSummary {
	checked := v.TotalRequests
	hasSuccess := v.Success2xx > 0
	has4xx := v.ClientError4xx > 0
	has5xx := v.ServerError5xx > 0
	hasNetwork := v.NetworkErrors > 0

	var msg string
	isReachable := false
	isHealthy := false
	isIndexable := false
	avgWords := 0
	if v.ParsedPages > 0 {
		avgWords = v.TotalWords / v.ParsedPages
	}

	switch mode {
	case UseCaseTrackBlogs:
		isReachable = hasSuccess && !has5xx && !hasNetwork
		switch {
		case checked == 0:
			msg = "We did not get any responses yet. Try again with a reachable blog URL."
		case hasNetwork:
			msg = "We could not reach this blog due to network or TLS errors."
		case has5xx:
			msg = "The blog server is returning 5xx errors, so it may be temporarily down."
		case has4xx && !hasSuccess:
			if v.LastStatusCode != 0 {
				msg = fmt.Sprintf("We only saw 4xx client errors (for example, %d %s). Check that the blog URL is correct.", v.LastStatusCode, http.StatusText(v.LastStatusCode))
			} else {
				msg = "We only saw 4xx client errors (like 404). Check that the blog URL is correct."
			}
		case hasSuccess && has4xx:
			msg = fmt.Sprintf("We reached %d page(s) successfully, but some returned 4xx errors.", v.Success2xx)
		default:
			msg = fmt.Sprintf("Your blog looks reachable. We saw %d successful page(s).", v.Success2xx)
		}
		if avgWords > 0 {
			msg += fmt.Sprintf(" Average content length is about %d words per page (largest page %d words).", avgWords, v.LongestPageWordCount)
		}

	case UseCaseSiteHealth:
		isHealthy = !has5xx && !hasNetwork
		switch {
		case checked == 0:
			msg = "No pages were checked yet, so we cannot rate site health."
		case hasNetwork:
			msg = "Network or TLS errors prevented a complete health check."
		case has5xx:
			if v.LastStatusCode != 0 {
				msg = fmt.Sprintf("We saw 5xx server errors (for example, %d %s). The site has availability issues.", v.LastStatusCode, http.StatusText(v.LastStatusCode))
			} else {
				msg = "We saw 5xx server errors. The site has availability issues."
			}
		case has4xx && !hasSuccess:
			if v.LastStatusCode != 0 {
				msg = fmt.Sprintf("Only 4xx client errors were seen (for example, %d %s). Many links may be broken or protected.", v.LastStatusCode, http.StatusText(v.LastStatusCode))
			} else {
				msg = "Only 4xx client errors were seen. Many links may be broken or protected."
			}
		case has4xx:
			if v.LastStatusCode != 0 {
				msg = fmt.Sprintf("Most pages responded, but some returned 4xx errors (for example, %d %s).", v.LastStatusCode, http.StatusText(v.LastStatusCode))
			} else {
				msg = "Most pages responded, but some returned 4xx errors (broken or restricted URLs)."
			}
		default:
			msg = "All checked pages responded without major server errors. The site looks healthy."
		}

	case UseCaseSearchIndex:
		isIndexable = hasSuccess && !hasNetwork
		switch {
		case checked == 0:
			msg = "We did not fetch any pages yet, so there is nothing to index."
		case hasNetwork:
			msg = "Network or TLS errors blocked us from gathering content to index."
		case hasSuccess:
			msg = fmt.Sprintf("We fetched %d page(s) successfully. They are good candidates for indexing.", v.Success2xx)
		default:
			msg = "We reached the site but did not see clear 2xx responses to index."
		}

	default:
		// Fallback for unknown/empty modes.
		switch {
		case checked == 0:
			msg = "No HTTP responses were recorded yet."
		case hasNetwork:
			msg = "Network or TLS errors occurred while crawling."
		case has5xx:
			if v.LastStatusCode != 0 {
				msg = fmt.Sprintf("Server responded with 5xx errors for some requests (for example, %d %s).", v.LastStatusCode, http.StatusText(v.LastStatusCode))
			} else {
				msg = "Server responded with 5xx errors for some requests."
			}
		case has4xx:
			if v.LastStatusCode != 0 {
				msg = fmt.Sprintf("Client-side 4xx errors were seen (for example, %d %s).", v.LastStatusCode, http.StatusText(v.LastStatusCode))
			} else {
				msg = "Client-side 4xx errors were seen (for example, 404 or 403)."
			}
		default:
			msg = "We saw successful responses without major errors."
		}
	}

	return ModeSummary{
		Mode:                 mode,
		CheckedPages:         checked,
		IsReachable:          isReachable,
		IsHealthy:            isHealthy,
		IsIndexable:          isIndexable,
		PrimaryStatus:        v.LastStatusCode,
		Message:              msg,
		RawStats:             v,
		AverageWordsPerPage:  avgWords,
		LongestPageURL:       v.LongestPageURL,
		LongestPageTitle:     v.LongestPageTitle,
		LongestPageWordCount: v.LongestPageWordCount,
	}
}
