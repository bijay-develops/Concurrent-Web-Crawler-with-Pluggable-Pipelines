# Per-Domain Backpressure Diagram

### ❌ Problem: One domain overloads everything

```
Parser
  │
  ▼
[ Global URL Queue ]
  │
  ▼
Workers
```

If `example.com` is slow:

* Workers block
* Other domains suffer ❌

---

### ✅ Solution: Per-domain queues + backpressure

```
                 ┌──────────────┐
                 │ example.com  │◄── bounded (10)
                 └──────┬───────┘
                        │
Parser ──► Router ──►   ├──────────────┐
                        │ google.com   │◄── bounded (20)
                        └──────┬───────┘
                               │
                               ├──────────────┐
                               │ github.com   │◄── bounded (5)
                               └──────┬───────┘
                                      │
                                   Workers
```

### What happens?

* `example.com` queue fills → **only that domain blocks**
* Other domains continue crawling ✅
* **Fairness + politeness**

---

### Key idea

> Backpressure is applied **per domain**, not globally.

---