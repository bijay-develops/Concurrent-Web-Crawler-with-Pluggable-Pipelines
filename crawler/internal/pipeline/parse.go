package pipeline

import (
	"context"
	"crawler/internal/shared"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

func ParseWorker(
	ctx context.Context,
	in <-chan shared.Item,
	out chan<- shared.Item,
	stats *shared.CrawlStats,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}
			if item.Response != nil {
				// Best-effort basic scraping/analytics for HTML responses.
				body, err := io.ReadAll(item.Response.Body)
				item.Response.Body.Close()
				item.Response = nil
				if err == nil && len(body) > 0 {
					title, wordCount, internalLinks, externalLinks, keywords := extractPageMetrics(item.URL, body)
					if stats != nil {
						stats.RecordPageMetrics(item.URL.String(), title, wordCount, internalLinks, externalLinks)
						stats.RecordTopics(title, keywords)
					}
				}
			}

			select {
			case out <- item:
			case <-ctx.Done():
				return
			}
		}
	}
}

// extractPageMetrics performs very lightweight HTML scraping to provide
// user-friendly analytics and keyword hints without external dependencies.
func extractPageMetrics(pageURL *url.URL, body []byte) (title string, wordCount, internalLinks, externalLinks int, keywords []string) {
	// Work on a lowercase copy for tag/attribute searches.
	lower := strings.ToLower(string(body))

	// 1. Title extraction: first <title>...</title> block.
	start := strings.Index(lower, "<title")
	if start >= 0 {
		// Skip to closing '>' of the <title> tag.
		endTagStart := strings.Index(lower[start:], ">")
		if endTagStart >= 0 {
			contentStart := start + endTagStart + 1
			end := strings.Index(lower[contentStart:], "</title>")
			if end >= 0 {
				title = strings.TrimSpace(string(body[contentStart : contentStart+end]))
			}
		}
	}

	// 2. Rough word count and per-page keyword candidates.
	noTags := stripHTMLTags(lower)
	if noTags != "" {
		words := strings.Fields(noTags)
		wordCount = len(words)
		keywords = topKeywords(words, 5)
	}

	// 3. Link counts: count internal vs external <a href="..."> links.
	internalLinks, externalLinks = countLinks(pageURL, lower)

	return title, wordCount, internalLinks, externalLinks, keywords
}

var tagRegexp = regexp.MustCompile(`<[^>]+>`)

func stripHTMLTags(s string) string {
	// Remove all angle-bracket tags; this is not a full HTML sanitizer
	// but works well enough for counting words.
	clean := tagRegexp.ReplaceAllString(s, " ")
	// Collapse whitespace.
	clean = strings.Join(strings.Fields(clean), " ")
	return clean
}

// Basic English stopwords to avoid boring topics.
var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "for": {}, "with": {}, "that": {}, "this": {}, "you": {}, "your": {}, "are": {}, "was": {}, "were": {}, "from": {}, "have": {}, "has": {}, "had": {}, "but": {}, "his": {}, "her": {}, "she": {}, "him": {}, "our": {}, "their": {}, "they": {}, "them": {}, "not": {}, "just": {}, "about": {}, "into": {}, "when": {}, "what": {}, "how": {}, "why": {}, "can": {}, "will": {}, "would": {}, "could": {}, "should": {}, "on": {}, "in": {}, "to": {}, "of": {}, "at": {}, "as": {}, "it": {}, "is": {}, "a": {}, "an": {},
}

func topKeywords(words []string, limit int) []string {
	if len(words) == 0 || limit <= 0 {
		return nil
	}
	counts := make(map[string]int)
	for _, w := range words {
		w = strings.Trim(w, " ,.!?:;\"'()[]{}<>")
		if w == "" {
			continue
		}
		lw := strings.ToLower(w)
		if len(lw) < 4 { // skip very short words
			continue
		}
		if _, skip := stopwords[lw]; skip {
			continue
		}
		counts[lw]++
	}
	if len(counts) == 0 {
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
	if len(items) > limit {
		items = items[:limit]
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.k)
	}
	return out
}

func countLinks(pageURL *url.URL, lowerHTML string) (internalLinks, externalLinks int) {
	if pageURL == nil {
		return 0, 0
	}
	// Very small regex to capture href values. It is intentionally
	// conservative and may miss some malformed cases.
	hrefRe := regexp.MustCompile(`href\s*=\s*"([^"]+)"`)
	matches := hrefRe.FindAllStringSubmatch(lowerHTML, -1)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		link := strings.TrimSpace(m[1])
		if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") {
			continue
		}

		u, err := url.Parse(link)
		if err != nil {
			continue
		}
		// Resolve relative links against the page URL.
		if !u.IsAbs() {
			u = pageURL.ResolveReference(u)
		}
		if equalHosts(u.Host, pageURL.Host) {
			internalLinks++
		} else {
			externalLinks++
		}
	}
	return internalLinks, externalLinks
}

func equalHosts(a, b string) bool {
	// Normalize trivial www. prefix differences.
	na := strings.TrimPrefix(a, "www.")
	nb := strings.TrimPrefix(b, "www.")
	return na == nb
}
