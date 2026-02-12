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

---
