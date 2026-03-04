package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"crawler/internal/crawler"
)

func main() {
	workers := flag.Int("workers", 8, "number of concurrent worker goroutines")
	maxDepth := flag.Int("depth", 2, "maximum crawl depth")
	seedURL := flag.String("url", "https://example.com", "seed URL to start crawling from")

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
	)

	if err := c.Run(ctx); err != nil {
		log.Println("crawler exited:", err)
	}
}
