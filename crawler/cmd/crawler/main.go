package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"crawler/internal/crawler"
)

func main() {
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
		crawler.WithWorkerCount(8),
		crawler.WithMaxDepth(3),
	)

	if err := c.Run(ctx); err != nil {
		log.Println("crawler exited:", err)
	}
}
