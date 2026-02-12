# Code Comparison (Side-by-Side)

## A. Single-Threaded Crawler

```go
func crawl(urls []string) {
    visited := make(map[string]bool)

    for len(urls) > 0 {
        url := urls[0]
        urls = urls[1:]

        if visited[url] {
            continue
        }

        visited[url] = true
        fmt.Println("Crawling:", url)

        links := extractLinks(url)
        urls = append(urls, links...)
    }
}
```

❌ Slow <br>
❌ No parallelism <br>
❌ Blocks on network I/O

---

## B. Go Concurrent Crawler

```go
var (
    visited = make(map[string]bool)
    mu      sync.Mutex
)

func worker(jobs <-chan string, wg *sync.WaitGroup) {
    for url := range jobs {
        mu.Lock()
        if visited[url] {
            mu.Unlock()
            wg.Done()
            continue
        }
        visited[url] = true
        mu.Unlock()

        fmt.Println("Crawling:", url)

        for _, link := range extractLinks(url) {
            wg.Add(1)
            jobs <- link
        }

        wg.Done()
    }
}

func main() {
    jobs := make(chan string)
    var wg sync.WaitGroup

    for i := 0; i < 3; i++ {
        go worker(jobs, &wg)
    }

    wg.Add(1)
    jobs <- "https://example.com"

    wg.Wait()
}
```

✅ Fast <br>
✅ Scalable <br>
✅ Thread-safe

---

