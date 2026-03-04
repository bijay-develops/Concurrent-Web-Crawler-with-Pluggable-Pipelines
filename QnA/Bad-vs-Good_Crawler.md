# Bad vs Good Crawler (Visual)

### ❌ BAD crawler

```
Parser ──► slice/list ──► Workers
           (unbounded)
```

What happens:

* URLs explode
* RAM explodes
* ❌ OOM crash

---

### ✅ GOOD crawler

```
Parser ──► [ bounded channel ] ──► Worker Pool
              ↑
         backpressure
```

What happens:

* Parser blocks when full
* Workers catch up
* ✅ Stable memory

---

### Side-by-side mental model

```
BAD:  Speed > Safety
GOOD: Throughput + Stability
```

In this project, the "good" crawler design is implemented using:

- Bounded channels between stages (see `internal/crawler` and `internal/pipeline`).
- Per-domain rate limiting in the fetch stage.
- Mode-aware stats and summaries so that you can quickly see how a crawl behaved.

For more Q&A on specific controls like **workers** and **max depth**, see [QnA/Workers_and_MaxDepth.md](QnA/Workers_and_MaxDepth.md).

---
