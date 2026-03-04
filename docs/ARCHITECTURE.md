# Architecture

## Overview
- Entry points:
  - Go CLI in `crawler/cmd/crawler/main.go`.
  - Web UI in `crawler/cmd/webui/main.go`.
  - JSON API server in `crawler/cmd/api/main.go`.
- Core crawler orchestration lives under `crawler/internal/crawler/`.
- Pluggable pipeline stages and infrastructure live under `crawler/internal/pipeline/`.
- Shared types (including the high-level use-case / mode and crawl statistics) live under `crawler/internal/shared/`.
- A small service layer for programmatic crawls lives under `crawler/internal/service/`.
- HTTP/JSON handlers live under `crawler/internal/httpapi/`.
- Simple file-based persistence for crawl summaries lives under `crawler/internal/store/`.

## Modules
- `cmd/crawler/main.go`
  - Sets up process context and signal handling.
  - Parses flags for worker count, max depth, seed URL, and a **mode** flag that selects one of three high-level use cases:
    - Track my favourite blogs
    - Internal Site Health Checker
    - Data Pipeline Search Index
  - Constructs a `Crawler` via functional options and starts it.
- `cmd/webui/main.go`
  - Serves a small HTML form for launching crawls from the browser.
  - Lets the user configure URL, workers, depth, and the same three modes via radio buttons.
  - Hosts the JSON API on the same server for the browser (the Web UI JS calls `/api/crawls`).
- `cmd/api/main.go`
  - Runs the same JSON API on a dedicated HTTP server (default `:8090`, configurable via `API_PORT`).
- `internal/crawler/`
  - Defines the main crawler type (`Crawler`) and its orchestration logic.
  - Provides options like `WithWorkerCount`, `WithMaxDepth`, `WithSeedURL`, and `WithUseCase` as referenced from the entrypoints.
  - Contains supporting types like `Schedular` and `WorkTracker`.
- `internal/shared/`
  - Defines shared data structures such as `Item`, which now carries the selected **use case** (`UseCaseTrackBlogs`, `UseCaseSiteHealth`, `UseCaseSearchIndex`).
  - Provides `CrawlStats` / `CrawlStatsView` and `ModeSummary` to aggregate and interpret crawl results.
- `internal/service/`
  - Wraps the crawler in a `CrawlService` with a simple API (`StartCrawl`) used by HTTP handlers and other integrations.
- `internal/httpapi/`
  - Maps HTTP requests to `CrawlService` calls.
  - Exposes endpoints such as `POST /api/crawls` (start a crawl and return stats/summary) and `GET /api/crawls/history` (list recent crawl summaries).
- `internal/store/`
  - Persists compact `CrawlRecord` summaries as JSON Lines under `crawler/data/crawls.jsonl`.
- `internal/pipeline/`
  - Hosts pluggable pipeline stages (`discover`, `fetch`, `filter`, `parse`, `store`) and related infrastructure (`interfaces`, `limiter`).
  - The current implementation focuses on wiring and rate-limiting primitives; most stage files are skeletal and document future responsibilities.

## High-Level Interactions
- The CLI, Web UI, and JSON API entrypoints configure and start the crawler through the internal package API and the `CrawlService`.
- The core crawler coordinates scheduling of crawl work, concurrency, and interaction with pipeline stages.
- A selected **mode** flows from the entrypoint into `internal/shared.Item` values and is used downstream (for example in the fetch stage and in result summaries) to adapt behavior and messaging.
- Conceptually, pipeline stages process crawl items in sequence (e.g., discover → fetch → filter → parse → store), with optional limiting; concrete logic is still being built out.
- After each crawl, aggregated stats and a human-readable `ModeSummary` are stored via the `store.FileStore` so history can be queried over HTTP.

## Mermaid System Diagram
```mermaid
flowchart TD
  CLI[cmd/crawler/main.go<br/>CLI entrypoint] --> CORE[internal/crawler<br/>Core crawler]
  API[cmd/api/main.go<br/>JSON API] --> SVC[internal/service<br/>CrawlService]
  WEBUI[cmd/webui/main.go<br/>Web UI + API] --> SVC
  SVC --> CORE
  CORE --> SCHED[internal/crawler/scheduler.go<br/>Scheduler]
  CORE --> PIPE[internal/pipeline<br/>Pipeline]
  PIPE --> DISC[discover.go<br/>Discover]
  PIPE --> FETCH[fetch.go<br/>Fetch]
  PIPE --> FILTER[filter.go<br/>Filter]
  PIPE --> PARSE[parse.go<br/>Parse]
  PIPE --> STORE[store.go<br/>Store]
  PIPE --> LIMIT[limiter.go<br/>Limiter]
  CORE --> STATS[internal/shared<br/>CrawlStats/ModeSummary]
  SVC --> PERSIST[internal/store<br/>FileStore (crawls.jsonl)]
```

## Notes
- The above diagram reflects the current package layout and function calls in `main.go` and `internal/crawler`.
- Some pipeline stage files are placeholders; their responsibilities are documented even where implementations are not yet complete.
