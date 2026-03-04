package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"crawler/internal/store"
)

func main() {
	filePath := flag.String("file", "data/crawls.jsonl", "path to crawls.jsonl history file")
	TopN := flag.Int("top", 10, "number of top topics to show")
	flag.Parse()

	if _, err := os.Stat(*filePath); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("history file not found: %s", *filePath)
		}
		log.Fatalf("cannot stat history file: %v", err)
	}

	fs := store.NewFileStore(*filePath)
	recs, err := fs.ListCrawls(context.Background())
	if err != nil {
		log.Fatalf("list crawls: %v", err)
	}
	if len(recs) == 0 {
		fmt.Println("No crawl history found.")
		return
	}

	// Aggregate topics across all crawls, applying an extra layer of noise
	// filtering so old CSS-like keywords do not dominate the report.
	topicCounts := make(map[string]int)
	topicExamples := make(map[string][]string)

	var largestPage *pageSummary
	var smallestPage *pageSummary

	for _, rec := range recs {
		stats := rec.Stats

		// Track largest and smallest pages by word count using the
		// per-crawl longest-page stats.
		if stats.LongestPageWordCount > 0 {
			ps := pageSummary{
				URL:   stats.LongestPageURL,
				Title: stats.LongestPageTitle,
				Words: stats.LongestPageWordCount,
			}
			if largestPage == nil || ps.Words > largestPage.Words {
				copy := ps
				largestPage = &copy
			}
			if smallestPage == nil || ps.Words < smallestPage.Words {
				copy := ps
				smallestPage = &copy
			}
		}

		// Aggregate topics from this crawl.
		for _, t := range stats.Topics {
			keyword := strings.TrimSpace(strings.ToLower(t.Keyword))
			if keyword == "" || isNoisyTopic(keyword) {
				continue
			}
			// Sum counts across crawls.
			topicCounts[keyword] += t.Count

			// Merge example titles, de-duplicated and capped.
			for _, ex := range t.ExampleTitles {
				if ex == "" {
					continue
				}
				existing := topicExamples[keyword]
				if len(existing) >= 3 {
					break
				}
				dup := false
				for _, e := range existing {
					if e == ex {
						dup = true
						break
					}
				}
				if !dup {
					topicExamples[keyword] = append(existing, ex)
				}
			}
		}
	}

	fmt.Printf("Analyzed %d crawl record(s) from %s\n\n", len(recs), *filePath)

	// Report top themes/topics.
	if len(topicCounts) == 0 {
		fmt.Println("No topics found in history (try running a crawl that parses content).")
	} else {
		fmt.Println("Top themes / topics:")
		type pair struct {
			k string
			c int
		}
		items := make([]pair, 0, len(topicCounts))
		for k, c := range topicCounts {
			items = append(items, pair{k: k, c: c})
		}
		sort.Slice(items, func(i, j int) bool {
			if items[i].c == items[j].c {
				return items[i].k < items[j].k
			}
			return items[i].c > items[j].c
		})
		if len(items) > *TopN {
			items = items[:*TopN]
		}
		for _, it := range items {
			examples := topicExamples[it.k]
			if len(examples) > 0 {
				fmt.Printf("- %s (%d) e.g. %s\n", it.k, it.c, strings.Join(examples, "; "))
			} else {
				fmt.Printf("- %s (%d)\n", it.k, it.c)
			}
		}
	}

	fmt.Println()

	// Report largest and smallest pages by word count.
	if largestPage == nil {
		fmt.Println("No page size information found.")
		return
	}

	fmt.Println("Largest page (by words):")
	fmt.Printf("- Title: %s\n", largestPage.Title)
	fmt.Printf("- URL:   %s\n", largestPage.URL)
	fmt.Printf("- Words: %d\n", largestPage.Words)

	if smallestPage != nil {
		fmt.Println()
		fmt.Println("Smallest page (by words):")
		fmt.Printf("- Title: %s\n", smallestPage.Title)
		fmt.Printf("- URL:   %s\n", smallestPage.URL)
		fmt.Printf("- Words: %d\n", smallestPage.Words)
	}
}

type pageSummary struct {
	URL   string
	Title string
	Words int
}

// isNoisyTopic applies extra filtering on top of the crawler's own
// keyword extraction so that historical CSS-like artifacts do not
// dominate the history analysis.
func isNoisyTopic(keyword string) bool {
	if keyword == "" {
		return true
	}

	// Drop obviously CSS-ish or layout-related tokens by pattern.
	if strings.HasPrefix(keyword, "kt-") || strings.HasPrefix(keyword, "kb-") || strings.Contains(keyword, "wp-block") {
		return true
	}

	// Discard tokens that contain characters other than letters and
	// simple spaces/hyphens. This filters out most CSS blobs such as
	// "1024px){.kt-inner-column-height-full...".
	hasLetter := false
	for _, ch := range keyword {
		if ch >= 'a' && ch <= 'z' {
			hasLetter = true
			continue
		}
		if ch == ' ' || ch == '-' {
			continue
		}
		return true
	}
	if !hasLetter {
		return true
	}

	// Explicitly drop a few known layout-related words that are not
	// interesting as content themes.
	switch keyword {
	case "none", "solid", "max-width", "margin", "padding", "column", "image", "img", "important":
		return true
	}

	return false
}
