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
			item.DiscoveredURLs = nil
			if item.Response != nil {
				// Only attempt HTML scraping/analytics for text/html-like
				// responses. This keeps us polite to APIs and binary
				// endpoints and avoids wasting work.
				ct := item.Response.Header.Get("Content-Type")
				if ct != "" && !strings.Contains(ct, "text/html") && !strings.Contains(ct, "application/xhtml+xml") {
					item.Response.Body.Close()
					item.Response = nil
				} else {
					// Best-effort basic scraping/analytics for HTML responses.
					body, err := io.ReadAll(item.Response.Body)
					item.Response.Body.Close()
					item.Response = nil
					if err == nil && len(body) > 0 {
						title, wordCount, internalLinks, externalLinks, keywords, discovered := extractPageMetrics(item.URL, body)
						item.DiscoveredURLs = discovered
						if stats != nil {
							stats.RecordPageMetrics(item.URL.String(), title, wordCount, internalLinks, externalLinks, keywords)
							stats.RecordTopics(title, keywords)
						}
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
func extractPageMetrics(pageURL *url.URL, body []byte) (title string, wordCount, internalLinks, externalLinks int, keywords []string, discovered []string) {
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
		discovered := extractInternalLinksRegex(pageURL, lower, 150)
		return title, wordCount, internalLinks, externalLinks, keywords, discovered
	}

	// Extract <title> from the full document. We do this separately
	// because our content walk intentionally skips <head>.
	title = findDocumentTitle(root)

	// Prefer extracting visible text from <main> or <article> when present.
	// This reduces noise from nav/sidebars on listing pages.
	contentRoot := firstElementByTag(root, "main")
	if contentRoot == nil {
		contentRoot = firstElementByTag(root, "article")
	}
	if contentRoot == nil {
		contentRoot = root
	}

	var textParts []string
	seenDiscovered := make(map[string]struct{})
	discovered = nil

	// Prefer post links found in <article> cards on listing pages.
	// This keeps discovery focused on real content and avoids scheduling
	// tag/author/meta links that often appear inside article cards.
	const maxDiscoveredPerPage = 150
	preferredLinks, preferredSet := extractPreferredArticleLinks(pageURL, contentRoot, maxDiscoveredPerPage)
	articleCardMode := len(preferredLinks) > 0
	for _, p := range preferredLinks {
		if p == "" {
			continue
		}
		if _, ok := seenDiscovered[p]; ok {
			continue
		}
		seenDiscovered[p] = struct{}{}
		discovered = append(discovered, p)
		if len(discovered) >= maxDiscoveredPerPage {
			break
		}
	}

	stopWalk := false

	// Walk the DOM once, collecting visible text and link hrefs.
	var visit func(n *html.Node)
	visit = func(n *html.Node) {
		if n == nil || stopWalk {
			return
		}
		if n.Type == html.ElementNode {
			name := strings.ToLower(n.Data)
			// Skip non-content sections entirely.
			switch name {
			case "script", "style", "noscript", "head", "header", "footer", "nav", "aside":
				return
			case "a":
				if pageURL != nil {
					for _, attr := range n.Attr {
						if !strings.EqualFold(attr.Key, "href") {
							continue
						}
						link := strings.TrimSpace(attr.Val)
						if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(strings.ToLower(link), "javascript:") {
							continue
						}
						u, err := url.Parse(link)
						if err != nil || u == nil {
							continue
						}
						if !u.IsAbs() {
							u = pageURL.ResolveReference(u)
						}
						if equalHosts(u.Host, pageURL.Host) {
							internalLinks++
							if shouldDiscoverURL(pageURL, u) {
								key := u.String()
								// If this anchor is inside an <article> card on a listing page,
								// only discover it when it is one of our preferred per-article
								// permalinks. This avoids tag/author/meta links within cards.
								if articleCardMode && hasAncestorTag(n, "article") {
									if _, ok := preferredSet[key]; !ok {
										continue
									}
								}
								if _, ok := seenDiscovered[key]; !ok {
									seenDiscovered[key] = struct{}{}
									discovered = append(discovered, key)
									if len(discovered) >= maxDiscoveredPerPage {
										stopWalk = true
										return
									}
								}
							}
						} else {
							externalLinks++
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
			if stopWalk {
				return
			}
		}
	}
	visit(contentRoot)

	joined := strings.ToLower(strings.Join(textParts, " "))
	joined = strings.Join(strings.Fields(joined), " ")
	if joined != "" {
		words := strings.Fields(joined)
		wordCount = len(words)
		keywords = topKeywords(words, 10)
	}

	return title, wordCount, internalLinks, externalLinks, keywords, discovered
}

func extractPreferredArticleLinks(pageURL *url.URL, root *html.Node, limit int) (links []string, preferredSet map[string]struct{}) {
	preferredSet = make(map[string]struct{})
	if pageURL == nil || root == nil || limit <= 0 {
		return nil, preferredSet
	}

	articles := findElementsByTag(root, "article", 80)
	// If there's only one <article>, it's usually a single post page.
	// In that case we should NOT treat it as a listing page.
	if len(articles) < 2 {
		return nil, preferredSet
	}

	for _, art := range articles {
		if len(links) >= limit {
			break
		}
		bestURL := ""
		bestScore := -1
		anchorsSeen := 0
		stop := false

		var walk func(*html.Node)
		walk = func(n *html.Node) {
			if n == nil || stop {
				return
			}
			if n.Type == html.ElementNode && strings.EqualFold(n.Data, "a") {
				anchorsSeen++
				if anchorsSeen > 80 {
					stop = true
					return
				}
				href := attrValue(n, "href")
				href = strings.TrimSpace(href)
				if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(strings.ToLower(href), "javascript:") {
					// ignore
				} else {
					u, err := url.Parse(href)
					if err == nil && u != nil {
						if !u.IsAbs() {
							u = pageURL.ResolveReference(u)
						}
						if equalHosts(u.Host, pageURL.Host) && shouldDiscoverURL(pageURL, u) {
							key := u.String()
							score := scoreArticleAnchor(n)
							// Small preference toward longer, more specific paths.
							score += strings.Count(strings.Trim(u.Path, "/"), "/")
							if score > bestScore {
								bestScore = score
								bestURL = key
							}
						}
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
				if stop {
					return
				}
			}
		}
		walk(art)

		if bestURL == "" {
			continue
		}
		if _, ok := preferredSet[bestURL]; ok {
			continue
		}
		preferredSet[bestURL] = struct{}{}
		links = append(links, bestURL)
	}

	return links, preferredSet
}

func findElementsByTag(root *html.Node, tag string, limit int) []*html.Node {
	if root == nil || tag == "" || limit <= 0 {
		return nil
	}
	out := make([]*html.Node, 0, 8)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil || len(out) >= limit {
			return
		}
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, tag) {
			out = append(out, n)
			if len(out) >= limit {
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
			if len(out) >= limit {
				return
			}
		}
	}
	walk(root)
	return out
}

func hasAncestorTag(n *html.Node, tag string) bool {
	if n == nil || tag == "" {
		return false
	}
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && strings.EqualFold(p.Data, tag) {
			return true
		}
	}
	return false
}

func attrValue(n *html.Node, key string) string {
	if n == nil || key == "" {
		return ""
	}
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func scoreArticleAnchor(a *html.Node) int {
	if a == nil {
		return 0
	}
	score := 0
	// Prefer heading links: <h2><a href=...>Title</a></h2>
	for p := a.Parent; p != nil; p = p.Parent {
		if p.Type != html.ElementNode {
			continue
		}
		name := strings.ToLower(p.Data)
		if name == "h1" || name == "h2" || name == "h3" {
			score += 60
			break
		}
		if name == "article" {
			break
		}
	}

	if rel := strings.ToLower(attrValue(a, "rel")); rel != "" {
		if strings.Contains(rel, "bookmark") {
			score += 30
		}
	}
	class := strings.ToLower(attrValue(a, "class"))
	id := strings.ToLower(attrValue(a, "id"))
	if strings.Contains(class, "entry-title") || strings.Contains(class, "post-title") || strings.Contains(id, "entry-title") || strings.Contains(id, "post-title") {
		score += 20
	}

	// Mild bump for non-trivial anchor text.
	if t := strings.TrimSpace(collectText(a)); len(t) >= 12 {
		score += 10
	}
	return score
}

// shouldDiscoverURL decides whether an internal link should be scheduled
// for crawling. It filters obvious non-HTML assets and self-links.
func shouldDiscoverURL(pageURL, u *url.URL) bool {
	if pageURL == nil || u == nil {
		return false
	}
	// Normalize obvious noise.
	u.Fragment = ""
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if !equalHosts(u.Host, pageURL.Host) {
		return false
	}
	// Skip most query-driven links; allow only tracking-only queries.
	if u.RawQuery != "" {
		if isOnlyTrackingQuery(u.RawQuery) {
			u.RawQuery = ""
		} else {
			return false
		}
	}
	// Avoid self-link loops.
	if u.String() == pageURL.String() {
		return false
	}
	path := strings.ToLower(u.Path)
	if path == "" || path == "/" {
		return false
	}
	path = strings.TrimSuffix(path, "/")
	if path == "" || path == "/" {
		return false
	}

	// Allow pagination within the same tag/archive listing page.
	if isAllowedPagination(pageURL.Path, u.Path) {
		return true
	}

	// Avoid non-content internal areas.
	if strings.Contains(path, "/wp-json/") || strings.Contains(path, "/wp-admin/") {
		return false
	}
	if strings.Contains(path, "/tag/") || strings.Contains(path, "/category/") || strings.Contains(path, "/author/") {
		return false
	}
	if strings.Contains(path, "/feed") || strings.Contains(path, "/login") || strings.Contains(path, "/subscribe") {
		return false
	}
	// Skip common asset/file extensions.
	switch {
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"), strings.HasSuffix(path, ".png"), strings.HasSuffix(path, ".gif"), strings.HasSuffix(path, ".svg"),
		strings.HasSuffix(path, ".css"), strings.HasSuffix(path, ".js"), strings.HasSuffix(path, ".json"), strings.HasSuffix(path, ".xml"),
		strings.HasSuffix(path, ".pdf"), strings.HasSuffix(path, ".zip"), strings.HasSuffix(path, ".gz"),
		strings.HasSuffix(path, ".mp3"), strings.HasSuffix(path, ".mp4"), strings.HasSuffix(path, ".webm"),
		strings.HasSuffix(path, ".woff"), strings.HasSuffix(path, ".woff2"), strings.HasSuffix(path, ".ttf"), strings.HasSuffix(path, ".eot"):
		return false
	}

	// Prefer crawling post-like URLs for better organization.
	return isLikelyPostPath(path)
}

func isAllowedPagination(fromPath, toPath string) bool {
	from := strings.TrimSuffix(strings.ToLower(fromPath), "/")
	to := strings.TrimSuffix(strings.ToLower(toPath), "/")
	if from == "" || to == "" {
		return false
	}
	// If we're on a tag page like /tag/calvinism, allow /tag/calvinism/page/2.
	if strings.Contains(from, "/tag/") {
		base := from
		if i := strings.Index(base, "/page/"); i >= 0 {
			base = base[:i]
		}
		return strings.HasPrefix(to, base+"/page/")
	}
	// Generic /page/N pagination from an archive page.
	if strings.Contains(from, "/page/") {
		base := from[:strings.Index(from, "/page/")]
		return strings.HasPrefix(to, base+"/page/")
	}
	return false
}

func isOnlyTrackingQuery(raw string) bool {
	q, err := url.ParseQuery(raw)
	if err != nil {
		return false
	}
	for k := range q {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "utm_") || lk == "fbclid" || lk == "gclid" {
			continue
		}
		return false
	}
	return len(q) > 0
}

func isLikelyPostPath(path string) bool {
	seg := strings.Split(strings.Trim(path, "/"), "/")
	if len(seg) == 0 {
		return false
	}

	// Date-based permalinks: /2026/03/slug
	for i := 0; i+1 < len(seg); i++ {
		if len(seg[i]) == 4 && isAllDigits(seg[i]) {
			return true
		}
	}

	// Common post permalinks on many blogs are 1 segment: /some-long-slug
	last := seg[len(seg)-1]
	if len(last) < 8 {
		return false
	}
	hasLetter := false
	for _, ch := range last {
		if ch >= 'a' && ch <= 'z' {
			hasLetter = true
			continue
		}
		if ch == '-' {
			continue
		}
		return false
	}
	return hasLetter
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func extractInternalLinksRegex(pageURL *url.URL, lowerHTML string, limit int) []string {
	if pageURL == nil || lowerHTML == "" || limit <= 0 {
		return nil
	}
	hrefRe := regexp.MustCompile(`href\s*=\s*"([^"]+)"`)
	matches := hrefRe.FindAllStringSubmatch(lowerHTML, -1)
	seen := make(map[string]struct{})
	out := make([]string, 0, 32)
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
		if !u.IsAbs() {
			u = pageURL.ResolveReference(u)
		}
		if !shouldDiscoverURL(pageURL, u) {
			continue
		}
		key := u.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
		if len(out) >= limit {
			break
		}
	}
	return out
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
	// common UI/navigation words that frequently dominate tag/archive pages
	"login": {}, "subscribe": {}, "newsletter": {}, "search": {}, "contact": {}, "donation": {}, "donate": {}, "tags": {}, "latest": {}, "home": {}, "public": {}, "members": {}, "member": {},
}

// firstElementByTag returns the first element node in the tree with the
// given lowercase tag name.
func firstElementByTag(root *html.Node, tag string) *html.Node {
	if root == nil || tag == "" {
		return nil
	}
	var found *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil || found != nil {
			return
		}
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, tag) {
			found = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
			if found != nil {
				return
			}
		}
	}
	walk(root)
	return found
}

// findDocumentTitle returns the content of the first <title> element.
func findDocumentTitle(root *html.Node) string {
	titleNode := firstElementByTag(root, "title")
	if titleNode == nil {
		return ""
	}
	return strings.TrimSpace(collectText(titleNode))
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
