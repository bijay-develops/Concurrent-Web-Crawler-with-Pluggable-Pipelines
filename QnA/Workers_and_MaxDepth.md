# Workers and Max Depth in the Concurrent Web Crawler

This Q&A explains two important knobs in the crawler: **Workers** and **Max depth**.

---

## Workers

**Meaning**

- How many crawl workers run in parallel.
- In this project, it’s the number of Go goroutines doing HTTP requests at the same time.

**What it does**

- Higher workers ⇒ more concurrent requests ⇒ faster coverage, but more load on:
  - The target site (more requests hitting it at once).
  - Your own machine and network.
- Lower workers ⇒ fewer parallel requests ⇒ slower but gentler on the site.
- Internally, each worker:
  - Pulls URLs from the scheduler.
  - Respects the per‑domain rate limiter.
  - Calls `GET` on the URL using the HTTP client.
  - Pushes the result (response + metadata) down the pipeline for further processing.

---

## Max depth

**Meaning**

- How far the crawler is allowed to go from the starting page in terms of link “hops”.
- Conceptually:
  - Depth `0` ⇒ only the seed URL.
  - Depth `1` ⇒ seed URL + the pages it links to.
  - Depth `2` ⇒ seed URL + its links + links from those pages, and so on.

**What it does**

- Acts as a **boundary** so the crawler doesn’t wander infinitely deep into a site.
- Each discovered URL is tagged with a `Depth` value.
  - If adding a new URL would exceed `maxDepth`, it is **not scheduled** for fetching.
- In the current code, the `DiscoverWorker` stage is still simple, but `maxDepth` is already wired in and will matter more as link extraction and recursive crawling grow.

In short:

- **Workers** control *how many things you do at once*.
- **Max depth** controls *how far you are allowed to go* from the starting URL.
