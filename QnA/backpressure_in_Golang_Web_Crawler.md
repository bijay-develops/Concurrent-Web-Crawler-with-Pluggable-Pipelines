## What is backpressure in a Go crawler?

**Backpressure** means:

> **Slowing down or blocking producers when consumers can’t keep up**

In a crawler:

* **Producers** → URL discoverers (parsers finding new links)
* **Consumers** → Fetchers / workers downloading pages

If producers keep generating URLs **faster than workers can fetch**, memory explodes and the crawler becomes unstable.

Backpressure **prevents overload**.

---

## Why backpressure is critical in crawlers

Without backpressure:

```
Parser → generates 1,000,000 URLs
Workers → can fetch only 100/sec
Queue → grows forever ❌
RAM → OOM crash
```

With backpressure:

```
Parser → BLOCKS when queue is full
Workers → drain queue
System → stays stable ✅
```

---

## Backpressure in Go (using channels)

### Bounded channel = built-in backpressure

```go
urlQueue := make(chan string, 1000) // bounded buffer
```

When full:

* `urlQueue <- url` **blocks**
* Producer slows down automatically

---

## Minimal crawler example

```go
func parser(out chan<- string, urls []string) {
	for _, u := range urls {
		out <- u // blocks if queue is full (backpressure)
	}
}

func worker(in <-chan string) {
	for url := range in {
		fetch(url)
	}
}
```

---

## Diagram (mental model)

```
[ Parser ]
    |
    |  (bounded channel)
    v
[ URL Queue ]  ←←← backpressure here
    |
    v
[ Fetch Workers ]
```

---

## Backpressure vs Rate Limiting

| Concept       | Purpose                                |
| ------------- | -------------------------------------- |
| Backpressure  | Protects **memory & system stability** |
| Rate limiting | Protects **target websites**           |
| Mutex         | Protects **shared data**               |
| Semaphore     | Limits **concurrency**                 |

All four are used together in a **production crawler**.

---

## 1-minute answer

> *Backpressure in a Go crawler is a mechanism that slows down URL producers when fetch workers can’t keep up. It’s usually implemented using bounded channels, so when the queue is full, producers block automatically. This prevents unbounded memory growth, stabilizes throughput, and allows the crawler to self-regulate under load.*

---

## Production patterns you should mention

* Bounded channels
* Worker pools
* Context cancellation
* Per-domain queues
* Persistent overflow (Redis / disk)
* Metrics (queue length, blocked time)

