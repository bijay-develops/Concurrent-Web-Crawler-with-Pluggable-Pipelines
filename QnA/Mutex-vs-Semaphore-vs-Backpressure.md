
# Mutex vs Semaphore vs Backpressure Diagram

### 🔒 Mutex (protects data)

```
        ┌──────────┐
        │ Goroutine│
        └────┬─────┘
             │ lock
             ▼
        ┌──────────┐
        │  MUTEX   │  ← only ONE allowed
        └──────────┘
```

Used for:

* Maps
* Counters
* Shared state

❌ Does NOT control load

---

### 🚦 Semaphore (limits concurrency)

```
Tokens = 5

G1 ─┐
G2 ─┼─► [ SEMAPHORE ] ─► Work
G3 ─┼─► (max 5)
G4 ─┼─►
G5 ─┘
G6 ❌ blocked
```

Used for:

* Max parallel HTTP requests
* CPU limits

---

### 🛑 Backpressure (controls flow)

```
Producer ──► [ bounded channel ] ──► Consumer
                 ↑
              BLOCK
```

Used for:

* Queue growth control
* Memory safety
* System stability

---

### 🔥 Comparison table (memorize this)

| Tool         | Controls          | Purpose       |
| ------------ | ----------------- | ------------- |
| Mutex        | Data access       | Correctness   |
| Semaphore    | Concurrency count | Load limiting |
| Backpressure | Flow rate         | Stability     |
| Rate limiter | Requests/sec      | Politeness    |

---

# 🧠 Final Mental Model (big picture)

```
Parser
  │
  ▼
[ Per-domain queues ]  ← backpressure
  │
  ▼
[ Worker pool ]        ← semaphore
  │
  ▼
[ Shared state ]       ← mutex
```

---

## 🎯 Closing line

> *A production crawler uses backpressure to control flow, semaphores to limit concurrency, and mutexes to protect shared state. Without backpressure, the crawler is unstable no matter how good the other controls are.*

---