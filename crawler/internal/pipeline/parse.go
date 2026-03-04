package pipeline

import (
	"context"
	"crawler/internal/shared"
	"io"
	"net/url"
	"regexp"
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
					title, wordCount, internalLinks, externalLinks := extractPageMetrics(item.URL, body)
					if stats != nil {
						stats.RecordPageMetrics(item.URL.String(), title, wordCount, internalLinks, externalLinks)
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
// user-friendly analytics without external dependencies.
func extractPageMetrics(pageURL *url.URL, body []byte) (title string, wordCount, internalLinks, externalLinks int) {
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

	// 2. Rough word count: strip tags and count fields.
	noTags := stripHTMLTags(lower)
	if noTags != "" {
		wordCount = len(strings.Fields(noTags))
	}

	// 3. Link counts: count internal vs external <a href="..."> links.
	internalLinks, externalLinks = countLinks(pageURL, lower)

	return title, wordCount, internalLinks, externalLinks
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
