package shared

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
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

	// Lightweight topic aggregation across parsed pages.
	topicCounts   map[string]int
	topicExamples map[string][]string

	// Per-page analytics (held in memory for the current crawl only).
	pages []PageRecord
}

// PageRecord captures per-page analytics used for exports and
// detailed inspection in UIs.
type PageRecord struct {
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	WordCount     int      `json:"wordCount"`
	InternalLinks int      `json:"internalLinks"`
	ExternalLinks int      `json:"externalLinks"`
	Keywords      []string `json:"keywords"`
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

	Topics []TopicSummary `json:"topics"`
}

// TopicSummary represents a simple theme/keyword discovered during a crawl.
type TopicSummary struct {
	Keyword       string   `json:"keyword"`
	Count         int      `json:"count"`
	ExampleTitles []string `json:"exampleTitles"`
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
	AverageWordsPerPage  int            `json:"averageWordsPerPage"`
	LongestPageURL       string         `json:"longestPageUrl"`
	LongestPageTitle     string         `json:"longestPageTitle"`
	LongestPageWordCount int            `json:"longestPageWordCount"`
	TopTopics            []TopicSummary `json:"topTopics"`
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
// successfully fetched page and stores a per-page record. It is safe
// for concurrent use.
func (s *CrawlStats) RecordPageMetrics(url, title string, wordCount, internalLinks, externalLinks int, keywords []string) {
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

	// Store a per-page record for systematic export. This remains in
	// memory for the life of a crawl and is not persisted to disk.
	rec := PageRecord{
		URL:           url,
		Title:         title,
		WordCount:     wordCount,
		InternalLinks: internalLinks,
		ExternalLinks: externalLinks,
		Keywords:      append([]string(nil), keywords...),
	}
	// Lazily allocate a slice on first use.
	// Use topicCounts map size as a rough guard against unbounded growth
	// in very large crawls; this project is designed for small demos.
	// We simply append here; callers reading page data should do so
	// before discarding the CrawlStats instance.
	//
	// Note: we intentionally do not cap the slice here to keep the
	// implementation simple for this educational project.
	if s.topicCounts == nil {
		// ensure maps exist if we haven't recorded topics yet
		s.topicCounts = make(map[string]int)
	}
	// Reuse topicExamples map to track that we've at least seen pages,
	// but page records themselves live only in memory.
	// We'll attach them via PagesSnapshot for export.
	// For now we don't need a dedicated slice field; use a hidden key.
	// To keep the struct straightforward, add a dedicated slice field.
	// (See PagesSnapshot implementation below.)
	//
	// Since Go doesn't support anonymous slice fields well with maps,
	// we'll maintain an internal slice on CrawlStats.
	s.pages = append(s.pages, rec)
}

// RecordTopics aggregates per-page keywords into crawl-level topic stats.
// It is safe for concurrent use.
func (s *CrawlStats) RecordTopics(title string, keywords []string) {
	if s == nil || len(keywords) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.topicCounts == nil {
		s.topicCounts = make(map[string]int)
	}
	if s.topicExamples == nil {
		s.topicExamples = make(map[string][]string)
	}

	for _, k := range keywords {
		if k == "" {
			continue
		}
		s.topicCounts[k]++
		// Keep up to 3 example titles per keyword for UI display.
		if title != "" {
			examples := s.topicExamples[k]
			if len(examples) < 3 {
				s.topicExamples[k] = append(examples, title)
			}
		}
	}
}

// Snapshot returns a lock-free copy of the stats for read-only use.
func (s *CrawlStats) Snapshot() CrawlStatsView {
	if s == nil {
		return CrawlStatsView{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	view := CrawlStatsView{
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

	// Derive a small, sorted list of top topics for reporting.
	if len(s.topicCounts) > 0 {
		view.Topics = TopTopicsFromMaps(s.topicCounts, s.topicExamples, 10)
	}

	return view
}

// PagesSnapshot returns a shallow copy of the per-page analytics slice
// so callers can safely inspect it without holding the lock.
func (s *CrawlStats) PagesSnapshot() []PageRecord {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.pages) == 0 {
		return nil
	}
	out := make([]PageRecord, len(s.pages))
	copy(out, s.pages)
	return out
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
		TopTopics:            v.Topics,
	}
}

// TopTopicsFromMaps converts raw keyword counts and examples into a sorted
// slice of TopicSummary limited to topN entries.
func TopTopicsFromMaps(counts map[string]int, examples map[string][]string, topN int) []TopicSummary {
	if len(counts) == 0 || topN <= 0 {
		return nil
	}
	type pair struct {
		k string
		c int
	}
	items := make([]pair, 0, len(counts))
	for k, c := range counts {
		items = append(items, pair{k: k, c: c})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].c == items[j].c {
			return items[i].k < items[j].k
		}
		return items[i].c > items[j].c
	})

	if len(items) > topN {
		items = items[:topN]
	}

	out := make([]TopicSummary, 0, len(items))
	for _, it := range items {
		out = append(out, TopicSummary{
			Keyword:       it.k,
			Count:         it.c,
			ExampleTitles: examples[it.k],
		})
	}
	return out
}
