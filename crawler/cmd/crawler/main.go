package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"crawler/internal/crawler"
	"crawler/internal/shared"
)

func main() {
	workers := flag.Int("workers", 8, "number of concurrent worker goroutines")
	maxDepth := flag.Int("depth", 2, "maximum crawl depth")
	seedURL := flag.String("url", "https://example.com", "seed URL to start crawling from")
	modeFlag := flag.String("mode", "blogs", "use case: blogs|health|search (1=blogs,2=health,3=search)")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		log.Println("Shutdown signal received, stopping crawler...")
		cancel()
	}()

	c := crawler.New(
		crawler.WithWorkerCount(*workers),
		crawler.WithMaxDepth(*maxDepth),
		crawler.WithSeedURL(*seedURL),
		crawler.WithUseCase(parseUseCase(*modeFlag)),
	)

	if err := c.Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			log.Println("crawler finished (context closed):", err)
		} else {
			log.Println("crawler exited with error:", err)
		}
	}
}

func parseUseCase(s string) shared.UseCase {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "1", "blogs", "blog", "track-blogs", "track my favourite blogs":
		return shared.UseCaseTrackBlogs
	case "2", "health", "site-health", "internal site health checker":
		return shared.UseCaseSiteHealth
	case "3", "search", "index", "search-index", "data pipeline search index":
		return shared.UseCaseSearchIndex
	default:
		return shared.UseCaseTrackBlogs
	}
}
