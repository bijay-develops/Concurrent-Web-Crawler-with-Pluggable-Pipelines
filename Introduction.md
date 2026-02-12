# Introduction

> A single-threaded crawler fetches one URL at a time, which is slow because network requests block execution. A Go concurrent crawler solves this by using goroutines to fetch multiple pages in parallel, channels to distribute URLs, and a mutex-protected shared map to avoid duplicate crawling. This improves performance, scalability, and safety, and allows controlled concurrency using worker pools and rate limiting.

