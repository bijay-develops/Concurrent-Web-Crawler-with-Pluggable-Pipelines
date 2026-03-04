# Data Flow

## Overview
- The project is structured as a concurrent web crawler with pluggable pipeline stages.
- Core control flow, scheduling, and concurrency wiring are implemented in `internal/crawler`.
- Pipeline stage files in `internal/pipeline` mostly describe intended behavior; their concrete logic is still evolving.

## End-to-End Flow (Conceptual)
- **Inputs**
  - Seed URLs provided via:
    - CLI flags (`-url`) in the `cmd/crawler` binary.
    - The Web UI form at `/` in `cmd/webui` (default URL prefilled for convenience).
    - JSON requests to `POST /api/crawls` in the API or Web UI (body includes `url`, `workers`, `depth`, `mode`).
  - A selected **mode** / use case (Track Blogs, Site Health, Search Index) provided by the user.
- **Processing**
  - The core crawler creates channels for each stage (seeds, scheduled, fetched, parsed, discovered).
  - A `Schedular` instance deduplicates URLs and forwards unique items.
  - Each `Item` carries the chosen **use case**, allowing workers and stores to log or adapt behavior per mode.
  - A pool of fetch workers (size controlled by `WithWorkerCount`) pulls from the scheduled channel and writes to the fetched channel, respecting per-domain rate limiting.
  - Downstream stages conceptually process items in sequence: discover → fetch → filter → parse → store, with limiting where appropriate and depth bounded by `maxDepth`.
  - During the fetch stage, a shared `CrawlStats` instance is updated with response counts (2xx/4xx/5xx/network errors, last status, etc.).
  - After the crawl completes, the stats are snapshotted into a `CrawlStatsView` and summarized into a `ModeSummary` (user-friendly text and booleans like `isHealthy`, `isReachable`, `isIndexable`).
- **Outputs**
  - Aggregated stats and summaries returned directly to callers:
    - CLI logs status, and the Web UI renders a "What this means" section.
    - API callers of `POST /api/crawls` receive `{ url, mode, stats, summary, error }` as JSON.
  - Persistent history:
    - Each finished crawl is saved as a compact `CrawlRecord` (URL, mode, stats, summary, error, timestamps) in a JSONL file at `crawler/data/crawls.jsonl` via the `internal/store` package.
    - `GET /api/crawls/history` reads this file and returns an array of historical crawl summaries.

## Mermaid Data Flow Diagram
```mermaid
flowchart LR
  CLI[CLI / Web UI / API] --> B[Core crawler<br/>(internal/crawler)]
  B --> C[Scheduler<br/>(scheduler.go)]
  C --> D[Discover stage<br/>(discover.go)]
  D --> E[Fetch stage<br/>(fetch.go)]
  E --> F[Filter stage<br/>(filter.go)]
  F --> G[Parse stage<br/>(parse.go)]
  G --> H[Store stage<br/>(store.go)]
  E --> STATS[Stats & Summary<br/>(internal/shared)]
  STATS --> HIST[History<br/>(internal/store, crawls.jsonl)]
  HIST --> API[HTTP JSON responses<br/>(internal/httpapi)]
```

## External Dependencies
- Standard library packages used in the entrypoint (`context`, `os/signal`, `syscall`, `time`, `log`).
- Standard library networking used in the pipeline (`net/http`).
- No external storage dependencies are defined in code yet; these will likely be introduced as the pipeline stages evolve.
