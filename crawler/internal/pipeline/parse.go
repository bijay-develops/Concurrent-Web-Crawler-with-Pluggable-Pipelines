package pipeline

import (
	"bytes"
	"context"
	"crawler/internal/shared"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
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
						stats.RecordPageMetrics(item.URL.String(), title, wordCount, internalLinks, externalLinks, keywords)
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

// extractPageMetrics performs HTML scraping using a DOM parser so that
// text and links are derived from the real structure instead of
// fragile regular expressions.
func extractPageMetrics(pageURL *url.URL, body []byte) (title string, wordCount, internalLinks, externalLinks int, keywords []string) {
	// Parse HTML into a node tree. If parsing fails, fall back to the
	// previous regex-based approach for robustness.
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil || root == nil {
		lower := strings.ToLower(string(body))
		noTags := stripHTMLTags(lower)
		if noTags != "" {
			words := strings.Fields(noTags)
			wordCount = len(words)
			keywords = topKeywords(words, 5)
		}
		internalLinks, externalLinks = countLinks(pageURL, lower)
		return title, wordCount, internalLinks, externalLinks, keywords
	}

	var textParts []string

	// Walk the DOM once, collecting title, visible text, and link hrefs.
	var visit func(n *html.Node)
	visit = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode {
			name := strings.ToLower(n.Data)
			// Skip non-content sections entirely.
			switch name {
			case "script", "style", "noscript", "head", "header", "footer", "nav", "aside":
				return
			case "title":
				if title == "" {
					if t := collectText(n); t != "" {
						title = t
					}
				}
			case "a":
				if pageURL != nil {
					for _, attr := range n.Attr {
						if strings.EqualFold(attr.Key, "href") {
							link := strings.TrimSpace(attr.Val)
							if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") {
								continue
							}
							u, err := url.Parse(link)
							if err != nil {
								continue
							}
							if !u.IsAbs() {
								u = pageURL.ResolveReference(u)
							}
							if equalHosts(u.Host, pageURL.Host) {
								internalLinks++
							} else {
								externalLinks++
							}
						}
					}
				}
			}
		}
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				textParts = append(textParts, text)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(root)

	joined := strings.ToLower(strings.Join(textParts, " "))
	joined = strings.Join(strings.Fields(joined), " ")
	if joined != "" {
		words := strings.Fields(joined)
		wordCount = len(words)
		keywords = topKeywords(words, 10)
	}

	return title, wordCount, internalLinks, externalLinks, keywords
}

var tagRegexp = regexp.MustCompile(`<[^>]+>`)

func stripHTMLTags(s string) string {
	clean := tagRegexp.ReplaceAllString(s, " ")
	clean = strings.Join(strings.Fields(clean), " ")
	return clean
}

// collectText returns all descendant text of a node joined with
// single spaces, ignoring nested script/style sections.
func collectText(n *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			name := strings.ToLower(node.Data)
			if name == "script" || name == "style" || name == "noscript" {
				return
			}
		}
		if node.Type == html.TextNode {
			if t := strings.TrimSpace(node.Data); t != "" {
				parts = append(parts, t)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(parts, " ")
}

// Basic English stopwords to avoid boring topics. This also includes
// a few layout/CSS-related words that often leak into text content
// and are not meaningful as themes (for example, "none", "solid",
// "important").
var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "for": {}, "with": {}, "that": {}, "this": {}, "you": {}, "your": {}, "are": {}, "was": {}, "were": {}, "from": {}, "have": {}, "has": {}, "had": {}, "but": {}, "his": {}, "her": {}, "she": {}, "him": {}, "our": {}, "their": {}, "they": {}, "them": {}, "not": {}, "just": {}, "about": {}, "into": {}, "when": {}, "what": {}, "how": {}, "why": {}, "can": {}, "will": {}, "would": {}, "could": {}, "should": {}, "on": {}, "in": {}, "to": {}, "of": {}, "at": {}, "as": {}, "it": {}, "is": {}, "a": {}, "an": {},
	// extra noise words we do not want as topics
	"none": {}, "solid": {}, "important": {},
}

func topKeywords(words []string, limit int) []string {
	if len(words) == 0 || limit <= 0 {
		return nil
	}

	// Normalize and filter tokens first.
	clean := make([]string, 0, len(words))
	for _, w := range words {
		w = strings.Trim(w, " ,.!?:;\"'()[]{}<>")
		if w == "" {
			continue
		}
		lw := strings.ToLower(w)
		if len(lw) < 4 { // skip very short words
			continue
		}
		// Skip any token that contains non-letter characters (digits,
		// punctuation, etc.) so we avoid CSS classes or IDs like
		// "kt-inner-columns" or "1024px".
		isAlpha := true
		for _, ch := range lw {
			if ch < 'a' || ch > 'z' {
				isAlpha = false
				break
			}
		}
		if !isAlpha {
			continue
		}
		if _, skip := stopwords[lw]; skip {
			continue
		}
		clean = append(clean, lw)
	}
	if len(clean) == 0 {
		return nil
	}

	counts := make(map[string]int)
	// Unigrams
	for _, tok := range clean {
		counts[tok]++
	}
	// Simple bigrams: "bible verses", "christian blog", etc.
	for i := 0; i+1 < len(clean); i++ {
		phrase := clean[i] + " " + clean[i+1]
		counts[phrase]++
	}

	// Require a minimum frequency so we only keep true themes.
	const minFreq = 2
	type pair struct {
		k string
		c int
	}
	items := make([]pair, 0, len(counts))
	for k, c := range counts {
		if c < minFreq {
			continue
		}
		items = append(items, pair{k: k, c: c})
	}
	if len(items) == 0 {
		return nil
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
