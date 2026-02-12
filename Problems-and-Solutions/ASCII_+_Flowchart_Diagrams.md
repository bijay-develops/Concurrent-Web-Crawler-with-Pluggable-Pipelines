# ASCII + Flowchart Diagrams

## A. Single-Threaded Crawler (Sequential)

### ASCII Diagram

```text
Start
  |
  v
+---------+
| Fetch A |
+---------+
  |
  v
+---------+
| Fetch B |
+---------+
  |
  v
+---------+
| Fetch C |
+---------+
  |
  v
 Done
```

### Flowchart View

```text
+-------+
| Start |
+-------+
    |
    v
+-------------+
| Fetch URL   |
+-------------+
    |
    v
+-------------+
| Extract     |
| Links       |
+-------------+
    |
    v
+-------------+
| Next URL?   |
+-------------+
    |
    v
  (Repeat)
```

❌ Network wait blocks entire program <br>
❌ Poor scalability

---

## B. Go Concurrent Crawler (Worker Pool)

### ASCII Diagram

```text
                +-------------------+
                |   URL Channel     |
                +-------------------+
                   |        |        |
          +--------+        |        +--------+
          |                 |                 |
+----------------+ +----------------+ +----------------+
| Worker 1 (go)  | | Worker 2 (go)  | | Worker 3 (go)  |
+----------------+ +----------------+ +----------------+
          |                 |                 |
          +---------+-------+--------+--------+
                    |
            +-------------------+
            | Visited Map 🔒     |
            | (Mutex Protected) |
            +-------------------+
```

### Flowchart View

```text
+-------+
| Start |
+-------+
    |
    v
+------------------+
| Push URL to Chan |
+------------------+
    |
    v
+------------------+
| Worker Pool      |
+------------------+
    |
    v
+------------------+
| Check Visited 🔒 |
+------------------+
    |
    v
+------------------+
| Fetch + Extract  |
+------------------+
    |
    v
+------------------+
| Push New URLs    |
+------------------+
```

✅ Parallel execution <br>
✅ Safe shared state <br>
✅ Controlled concurrency

---