# How to Rate-Limit Concurrent Crawlers

Rate limiting prevents:

* Overloading websites
* Getting blocked
* Breaking terms of service

---

## A. Rate Limiting Using `time.Ticker` (Most Common)

### Example: 5 requests per second

```go
rateLimiter := time.NewTicker(200 * time.Millisecond)

func worker(jobs <-chan string) {
    for url := range jobs {
        <-rateLimiter.C   // wait before request
        fetch(url)
    }
}
```

✔ Simple
✔ Effective

---

## B. Token Bucket Pattern (Channel-Based)

```go
tokens := make(chan struct{}, 5)

// refill tokens
go func() {
    ticker := time.NewTicker(time.Second)
    for range ticker.C {
        for i := 0; i < 5; i++ {
            tokens <- struct{}{}
        }
    }
}()

func worker(jobs <-chan string) {
    for url := range jobs {
        <-tokens   // acquire token
        fetch(url)
    }
}
```

✔ Burst control
✔ Precise rate limiting

---

## C. Per-Domain Rate Limiting (Advanced)

```text
example.com → 2 req/sec
api.site.com → 1 req/sec
```

Use:

* Map[domain]*rateLimiter
* Mutex to protect map

Used by **search engines** and **large scrapers**.

---

# Final Summary (Strong Closing)

* **Single-threaded crawlers** are simple but slow
* **Concurrent crawlers** maximize CPU and network usage
* **Mutexes** protect shared state
* **Channels** coordinate work
* **Rate limiting** ensures ethical, safe crawling

---