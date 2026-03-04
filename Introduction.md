# Introduction

> A single-threaded crawler fetches one URL at a time, which is slow because network requests block execution. A Go concurrent crawler solves this by using goroutines to fetch multiple pages in parallel, channels to distribute URLs, and a mutex-protected shared map to avoid duplicate crawling. This improves performance, scalability, and safety, and allows controlled concurrency using worker pools and rate limiting.

## 1) What this tool can do
From a developer’s perspective, this project is a small-but-real concurrent crawler you can run as a CLI or as a local Web UI/JSON API.

- **Crawl a site concurrently** using a worker pool and a channel-based pipeline.
- **Deduplicate and cap work** via the scheduler (unique-URL dedupe + a conservative max unique pages cap per crawl).
- **Fetch politely per-domain** (per-domain spacing) and handle common transient failures (e.g., backoff on `429 Too Many Requests`).
- **Parse HTML pages into analytics** (best-effort): title, word count, internal/external link counts, and lightweight keywords/topics.
- **Discover new internal links** from parsed pages and feed them back into the scheduler until `maxDepth` is reached.
- **Expose results in multiple ways**:
	- Web UI at `:8080` to run a crawl interactively.
	- JSON API endpoints (used by the Web UI) to start crawls and read history.
	- Download/export helpers in the UI (JSON/CSV/TXT) for page-level analytics.
- **Persist crawl summaries** as JSON Lines (`data/crawls.jsonl`) and show recent history.

## 2) Limitations of the tool
This is intentionally conservative and demo-friendly rather than a “crawl the entire internet” system.

- **Not a full web-rendering crawler**: it does not execute JavaScript or render dynamic SPA content.
- **Discovery is intentionally conservative**:
	- It focuses on same-host links.
	- It applies heuristics to avoid noisy/non-content URLs.
	- The crawler caps unique pages per run by default to prevent accidental large crawls.
- **Parsing is best-effort**: only HTML-like responses are parsed; non-HTML content is skipped.
- **No robots.txt enforcement (yet)**: you should still respect site policies and terms of service when crawling.
- **Not distributed**: runs as a single process with in-memory scheduling and stats for a given crawl.
- **Persistence is summary-level**: history stores compact crawl summaries; it does not persist a full crawl graph or full page content.

## 3) Future possible plan
If you want to evolve this into something closer to production tooling, these are natural next steps (and they align with the existing “pluggable pipeline” layout):

- **Robots/sitemaps support**: optional robots.txt compliance + sitemap ingestion to seed discovery.
- **Stronger URL normalization**: canonicalization rules, configurable query handling, and better duplicate collapsing.
- **Pluggable filter policies**: allow/deny rules by path, content type, status code, depth, or keyword/topic.
- **Storage stage integration**: wire a real store stage to persist page-level analytics (DB/file/object storage) in addition to crawl summaries.
- **Per-domain concurrency controls**: enforce both “requests per second” and “max concurrent requests per domain”.
- **Better observability**: structured logs, timing metrics per stage, and trace IDs across the crawl.
- **Distributed scheduling** (optional): move the scheduler/queue to Redis/Kafka/Postgres for multi-node crawling.

