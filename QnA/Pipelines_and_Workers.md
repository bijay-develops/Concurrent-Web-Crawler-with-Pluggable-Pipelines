# Pipelines and Workers in the Crawler (Explained Like I’m 12)

## Question

How do the worker "robots" in this project run through pipelines like:

- fetch (download the page)
- parse (understand what’s on it)
- filter (decide if it’s interesting)
- store (save what they found)

## Short Answer

The crawler is built like a small factory line.
Each page (URL) is a box that moves through different machines:

1. **FetchWorker** – goes to the website and downloads the page.
2. **ParseWorker** – looks at the page body and processes it.
3. **Filter** – decides if we want to keep this page or ignore it.
4. **Store** – saves the pages we decided to keep.

The boxes move between machines using Go channels, like pipes.

---

## The Factory Line Picture

Imagine you have a box that has only a URL written on it:

```text
[ Start URL ]
      |
      v
```

### 1. Fetch: Download the page

The first machine is the **FetchWorker**.
It takes the URL, sends an HTTP request, and gets back a response (the page).

File: `crawler/internal/pipeline/fetch.go`

What it does (simplified):

```text
[ Start URL ]
      |
      v
 [ FetchWorker ]
   - Makes HTTP request
   - Gets the response (HTML)
      |
      v
```

After this step, the box now contains:

- The URL
- The HTTP response (what the server sent back)

### 2. Parse: Understand the page

Next, the box goes to the **ParseWorker**.
In this project, the parsing step is simple: it mainly handles the response body.

File: `crawler/internal/pipeline/parse.go`

Idea:

```text
[ FetchWorker ]
      |
      v
 [ ParseWorker ]
   - Reads/handles the page body
   - (In a bigger project: find links, text, titles, etc.)
      |
      v
```

Now the box has a “processed” version of the page.

### 3. Filter: Is this page interesting?

Then the box goes past a **Filter**.

File: `crawler/internal/pipeline/filter.go`

Right now, the filter is `AllowAllFilter`, which always returns `true`:

- That means: *“Let everything pass, keep all pages.”*

In a more advanced filter you could:

- Only allow pages from a certain domain.
- Only keep pages where the URL contains `blog`.

Picture:

```text
[ ParseWorker ]
      |
      v
   [ Filter ]
   - Ask: Allow(item)?
   - true  -> keep it
   - false -> drop it
      |
      v
 (only allowed items go forward)
```

### 4. Store: Save what we found

Finally, the box reaches the **Store** step.

File: `crawler/internal/pipeline/store.go`

There is a `LogStore` type with a `Store` method.
Right now it just logs:

- `stored: <url>`

This is a very simple way to “save” results.
In a real system this might:

- Write to a file.
- Save into a database.
- Add to a search index.

Picture:

```text
   [ Filter ]
      |
      v
   [ Store ]
   - Save or log the item
```

---

## Full Pipeline Overview

Putting it all together:

```text
Start URL
   |
   v
FetchWorker  ->  ParseWorker  ->  Filter  ->  Store
   |             |                |          |
 download       understand       decide     save
 page           page             keep?      it
```

- Each step is in its own file under `crawler/internal/pipeline/`.
- Workers use **channels** (`in` and `out`) to pass items along.
- This is why we call it a **pipeline**: data flows through steps, one after another.

---

## Channel Wiring: How the “Pipes” Connect in Code

So where do we actually connect all these machines with pipes?
That happens in the crawler’s `Run` method.

Main file: [crawler/internal/crawler/crawler.go](crawler/internal/crawler/crawler.go)

### 1. Pipes (channels) being created

Inside `Run`:

```go
seeds := make(chan shared.Item)
scheduled := make(chan shared.Item)
fetched := make(chan shared.Item)
parsed := make(chan shared.Item)
discovered := make(chan shared.Item)
```

These are our pipes:

- `seeds`      – where the very first URL goes.
- `scheduled`  – items approved by the scheduler.
- `fetched`    – items after downloading the page.
- `parsed`     – items after parsing.
- `discovered` – items after discovery.

### 2. Scheduler: seeds → scheduled (and discovered → scheduled)

```go
scheduler := NewScheduler()

go scheduler.Schedule(ctx, seeds, scheduled)
go scheduler.Schedule(ctx, discovered, scheduled)
```

Picture:

```text
[ seeds ] -----> (Scheduler) -----> [ scheduled ]
                           ^
                           |
                  [ discovered ]
```

The scheduler:

- Reads from `seeds` and `discovered`.
- Skips URLs it has already seen (no duplicates).
- Sends unique items into `scheduled`.

### 3. Fetch workers: scheduled → fetched

```go
client := pipeline.NewHTTPClient(10 * time.Second)
limiter := pipeline.NewDomainLimiter(500 * time.Millisecond)

for i := 0; i < c.workers; i++ {
      go func() {
            defer fetchWG.Done()
            pipeline.FetchWorker(ctx, client, limiter, scheduled, fetched, c.mode, c.stats)
      }()
}
```

Many fetch workers run at the same time:

```text
[ scheduled ] --> (FetchWorker x N) --> [ fetched ]
```

Each worker:

- Reads an item from `scheduled`.
- Makes an HTTP request.
- Puts the result into `fetched`.

### 4. Parse worker: fetched → parsed

```go
go pipeline.ParseWorker(ctx, fetched, parsed)
```

Picture:

```text
[ fetched ] --> (ParseWorker) --> [ parsed ]
```

The parser handles the response body and passes the item along.

### 5. Discover worker: parsed → discovered

```go
go pipeline.DiscoverWorker(ctx, parsed, discovered, c.maxDepth, tracker)
```

In [crawler/internal/pipeline/discover.go](crawler/internal/pipeline/discover.go), `DiscoverWorker`:

- Reads from `parsed`.
- Calls `tracker.Done()` when a URL is finished.
- Checks the depth (stops if `Depth >= maxDepth`).
- Would send newly found URLs into `discovered` (placeholder for now).

Picture:

```text
[ parsed ] --> (DiscoverWorker) --> [ discovered ]
```

Then `discovered` goes back into the scheduler, creating a loop:

```text
Seed URL
   |
   v
[ seeds ] --> Scheduler --> [ scheduled ]
                               |
                               v
                     FetchWorker(s) --> [ fetched ] --> ParseWorker --> [ parsed ]
                                                                                                          |
                                                                                                          v
                                                                                              DiscoverWorker --> [ discovered ]
                                                                                                          |
                                                                                                          v
                                                                                                Scheduler (again)
```

This is the full “assembly line” built with channels and goroutines.

## Why This Is Cool

- You can swap parts easily (e.g., replace `AllowAllFilter` with a smarter filter).
- You can add new steps (like "extract links" or "clean text").
- Many workers can run at the same time (concurrency), so the crawler can be fast.

Now when you look at the code, you can match each pipe (channel) and machine (worker) to this diagram and see how the whole crawler factory works.