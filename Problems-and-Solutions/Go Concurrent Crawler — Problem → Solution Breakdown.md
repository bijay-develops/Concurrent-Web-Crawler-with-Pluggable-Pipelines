A **Go Concurrent Crawler** solves several **real and common problems** that appear when you try to crawl the web **efficiently and safely**.

Below is a **problem → solution mapping**, with concrete examples.

---

## 1. Slow crawling (sequential execution)

### ❌ Problem

If you crawl URLs **one by one**, the program:

* Waits for each network request
* Wastes CPU time
* Becomes very slow for large websites

```text
Fetch A → wait → Fetch B → wait → Fetch C
```

### ✅ Solution: Concurrency (Goroutines)

The crawler fetches **many pages at the same time**.

```text
Fetch A ─┐
Fetch B ─┼─ concurrently
Fetch C ─┘
```

➡ **Result:** Much faster crawling.

---

## 2. Race conditions on shared data

### ❌ Problem

Multiple goroutines access a shared `visited` map:

```go
visited[url] = true // unsafe
```

This can cause:

* Crashes
* Corrupted map
* Duplicate crawling

### ✅ Solution: Mutex / thread-safe access

```go
mu.Lock()
visited[url] = true
mu.Unlock()
```

➡ **Result:** Safe, predictable behavior.

---

## 3. Duplicate URL crawling

### ❌ Problem

Same link appears on multiple pages:

```text
/home → /about
/blog → /about
```

Without tracking:

* `/about` is crawled multiple times
* Wasted bandwidth and time

### ✅ Solution: Shared “visited” map

```go
if visited[url] {
    return
}
```

➡ **Result:** Each URL is crawled only once.

---

## 4. Uncontrolled goroutine creation

### ❌ Problem

Naively spawning goroutines:

```go
go crawl(link)
```

Can cause:

* Thousands of goroutines
* Memory exhaustion
* Program crash

### ✅ Solution: Worker pool

Limit concurrency with fixed workers.

```go
for i := 0; i < 10; i++ {
    go worker(jobs)
}
```

➡ **Result:** Controlled resource usage.

---

## 5. Deadlocks & coordination issues

### ❌ Problem

Main function exits before goroutines finish:

```go
go crawl(url)
main exits ❌
```

### ✅ Solution: WaitGroup

```go
wg.Add(1)
go crawl(url)
wg.Wait()
```

➡ **Result:** Program waits until crawling is complete.

---

## 6. Inefficient communication between workers

### ❌ Problem

Using shared variables for communication:

* Hard to reason about
* Error-prone

### ✅ Solution: Channels

Channels act as **safe queues**.

```go
jobs <- url
url := <-jobs
```

➡ **Result:** Clean, safe coordination.

---

## 7. Network-bound inefficiency

### ❌ Problem

Network calls block threads:

* CPU sits idle
* Poor utilization

### ✅ Solution: Go’s lightweight goroutines

* Goroutines are cheap
* Thousands can run efficiently

➡ **Result:** High throughput crawling.

---

## 8. Infinite crawling loops

### ❌ Problem

Pages link back to each other:

```text
A → B → A → B → ...
```

### ✅ Solution

* Visited map
* Optional depth limit

```go
if depth > maxDepth {
    return
}
```

➡ **Result:** Controlled crawl boundaries.

---

## 9. Scalability problems

### ❌ Problem

Sequential crawlers don’t scale with:

* Large websites
* Many domains

### ✅ Solution

* Concurrency
* Worker pools
* Channels

➡ **Result:** Scales with CPU and network capacity.

---

## 10. Interview / System Design Value

This project demonstrates your ability to:

* Handle **concurrency**
* Avoid **race conditions**
* Design **scalable systems**
* Use Go idioms correctly

---

## Summary Table

| Problem              | How crawler solves it |
| -------------------- | --------------------- |
| Slow crawling        | Goroutines            |
| Race conditions      | Mutex                 |
| Duplicate URLs       | Shared visited map    |
| Too many goroutines  | Worker pool           |
| Program exits early  | WaitGroup             |
| Unsafe communication | Channels              |
| Infinite loops       | Visited + depth       |
| Poor scalability     | Concurrent design     |

---
