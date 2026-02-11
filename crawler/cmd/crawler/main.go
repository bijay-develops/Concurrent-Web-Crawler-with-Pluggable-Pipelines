package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crawler/internal/crawler"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %s, shutting down", sig)
		cancel()
	}()

	cfg := crawler.Config{
		WoekerCount: 8,
	}

	c := crawler.New(cfg)

	if err := c.Run(ctx); err != nil {
		log.Printf("crawler exited with error: %v", err)
	}

	//optional: bounded wait for cleanup
	time.Sleep(100 * time.Millisecond)
}
