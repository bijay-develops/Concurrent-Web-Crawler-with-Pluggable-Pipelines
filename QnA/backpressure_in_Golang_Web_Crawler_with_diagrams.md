## 1️⃣ Crawler WITHOUT backpressure (❌ problem)

```
        ┌──────────┐
        │  Parser  │
        │ (finds   │
        │  links)  │
        └────┬─────┘
             │
             │  keeps sending URLs
             ▼
      ┌──────────────┐
      │ URL Queue    │  ← UNBOUNDED
      │ (slice/list) │
      └──────────────┘
             │
             ▼
      ┌──────────────┐
      │ Fetch Workers│
      │ (slow I/O)   │
      └──────────────┘
```

### What goes wrong?

* Parser is **fast**
* Fetching is **slow**
* Queue grows **forever**
* ❌ Memory leak → crash

---

## 2️⃣ Crawler WITH backpressure (✅ correct)

```
        ┌──────────┐
        │  Parser  │
        └────┬─────┘
             │   send URL
             │
             ▼
     ┌──────────────────┐
     │ BOUNDED CHANNEL  │  size = 1000
     │  (urlQueue)     │  ← BACKPRESSURE
     └──────────────────┘
             │
             ▼
      ┌──────────────┐
      │ Fetch Workers│
      │ (N goroutines)│
      └──────────────┘
```

### What happens now?

* Channel gets **full**
* `urlQueue <- url` **BLOCKS**
* Parser is forced to **wait**
* Workers catch up
* ✅ System stays stable

---

## 3️⃣ Zoom-in: how blocking works

```
Parser tries to send URL
          │
          ▼
  urlQueue <- url
          │
          │ (queue full)
          ▼
     ┌───────────┐
     │  BLOCKED  │  ← backpressure
     └───────────┘
          │
          │ worker consumes URL
          ▼
     Parser resumes
```

👉 **Blocking = backpressure**

No sleeps, no hacks — Go does it naturally.

---

## 4️⃣ Real-life analogy 🚰

```
Water Tank (Channel)
Capacity = 1000 liters

Tap (Parser)        Outlet (Workers)
   │                     │
   ▼                     ▼
┌────────────────────────────┐
│            TANK            │
└────────────────────────────┘
```

* Tank full → tap **cannot pour**
* Outlet drains → tap resumes

This is **backpressure**.

---

## 5️⃣ Code + diagram together

```go
urlQueue := make(chan string, 1000) // bounded

go parser(urlQueue)
go worker(urlQueue)
```

```
parser ──► [ channel (1000) ] ──► workers
           ↑
        backpressure
```

---

## 6️⃣ Why interviewers love this concept

Backpressure proves you understand:

* Concurrency
* System stability
* Memory safety
* Real-world production systems

---