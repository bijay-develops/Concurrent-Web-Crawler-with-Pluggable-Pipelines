# 1. Single-Threaded vs Go Concurrent Crawler

## A. Execution Model Comparison

| Aspect           | Single-Threaded Crawler    | Go Concurrent Crawler     |
| ---------------- | -------------------------- | ------------------------- |
| Execution        | One URL at a time          | Many URLs in parallel     |
| Speed            | Slow (network waits block) | Fast (overlaps I/O waits) |
| CPU Utilization  | Poor                       | High                      |
| Scalability      | Low                        | High                      |
| Complexity       | Simple                     | Moderate                  |
| Resource Control | Implicit                   | Explicit (worker pool)    |
| Fault Isolation  | Low                        | Higher                    |

---

## B. Behavior Comparison (Example)

### Scenario

Crawl 4 pages, each taking **1 second** to fetch.

### Single-Threaded

```text
Time →
[ A ] [ B ] [ C ] [ D ] = 4 seconds
```

### Concurrent (4 workers)

```text
Time →
[ A ]
[ B ]
[ C ]
[ D ] = ~1 second
```

➡ **Same work, ~4× faster**

---

