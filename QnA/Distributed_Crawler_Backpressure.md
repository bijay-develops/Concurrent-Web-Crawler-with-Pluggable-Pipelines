# Distributed Crawler Backpressure

### ❌ Naive distributed crawler

```
Node A → Redis Queue ← Node B
             ↑
         infinite push ❌
```

* Redis grows unbounded
* Nodes overwhelm shared storage

---

### ✅ Distributed backpressure (correct)

```
             ┌───────────────────┐
             │  Redis / Kafka    │
             │  (bounded queues) │◄── BACKPRESSURE
             └───────┬───────────┘
                     │
      ┌──────────────┼──────────────┐
      ▼              ▼              ▼
  Worker A        Worker B        Worker C
```

### How backpressure works here:

* Queue size limit reached
* Producers **block / slow / reject**
* Nodes auto-throttle

---

### Production patterns

* Redis `BLPOP`
* Kafka consumer lag
* Token buckets per domain
* Crawl budget enforcement

---

### One-liner (interview gold)

> *Distributed backpressure prevents fast nodes from overwhelming shared queues and keeps the cluster stable.*

---