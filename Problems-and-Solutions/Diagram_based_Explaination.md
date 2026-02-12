# 2. Diagram-Based Explanation

## A. Single-Threaded Crawler (Sequential)

```text
+--------+
| Start  |
+--------+
     |
     v
+----------+
| Fetch A |
+----------+
     |
     v
+----------+
| Fetch B |
+----------+
     |
     v
+----------+
| Fetch C |
+----------+
     |
     v
+--------+
| Done   |
+--------+
```

### Problems

* Idle CPU during network waits
* Cannot scale
* Inefficient for large sites

---

## B. Go Concurrent Crawler (Worker Pool)

```text
                +----------------+
                |   URL Queue    |  (Channel)
                +----------------+
                    |    |    |
         +-----------+    |    +-----------+
         |                |                |
+---------------+ +---------------+ +---------------+
|   Worker 1    | |   Worker 2    | |   Worker 3    |
|  (goroutine)  | |  (goroutine)  | |  (goroutine)  |
+---------------+ +---------------+ +---------------+
         |                |                |
         +--------+-------+--------+-------+
                  |
          +------------------+
          |  Visited Map 🔒  | (Mutex)
          +------------------+
```

### Benefits

* Parallel fetching
* Controlled concurrency
* Safe shared state
* High throughput

---

## C. Data Safety Diagram

```text
Goroutine A ──┐
              ├── Lock ──> Visited Map ──> Unlock
Goroutine B ──┘
```

✔ Prevents race conditions
✔ Ensures each URL is crawled once

---

