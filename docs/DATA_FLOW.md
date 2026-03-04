# Data Flow

## Overview
- The project is structured as a concurrent web crawler with pluggable pipeline stages.
- Core control flow, scheduling, and concurrency wiring are implemented in `internal/crawler`.
- Pipeline stage files in `internal/pipeline` mostly describe intended behavior; their concrete logic is still evolving.

## End-to-End Flow (Conceptual)
- **Inputs**
  - Seed URLs provided via CLI flags or the Web UI form (by default `https://example.com`).
  - A selected **mode** / use case (Track Blogs, Site Health, Search Index) provided by the user.
  - **Processing**
  - The core crawler creates channels for each stage (seeds, scheduled, fetched, parsed, discovered).
  - A `Schedular` instance deduplicates URLs and forwards unique items.
  - Each `Item` carries the chosen **use case**, allowing workers and stores to log or adapt behavior per mode.
  - A pool of fetch workers (size controlled by `WithWorkerCount`) pulls from the scheduled channel and writes to the fetched channel, respecting per-domain rate limiting.
  - Downstream stages conceptually process items in sequence: discover → fetch → filter → parse → store, with limiting where appropriate and depth bounded by `maxDepth`.
- **Outputs**
  - Discovered items and any stored results (the storage format/destination is not yet implemented).

## Mermaid Data Flow Diagram
```mermaid
flowchart LR
  A[Seed URLs / input jobs] --> B[Core crawler<br/>(internal/crawler)]
  B --> C[Scheduler<br/>(scheduler.go)]
  C --> D[Discover stage<br/>(discover.go)]
  D --> E[Fetch stage<br/>(fetch.go)]
  E --> F[Filter stage<br/>(filter.go)]
  F --> G[Parse stage<br/>(parse.go)]
  G --> H[Store stage<br/>(store.go)]
  H --> I[Limiter<br/>(limiter.go, optional)]
  I --> J[Outputs<br/>(stored results)]
```

## External Dependencies
- Standard library packages used in the entrypoint (`context`, `os/signal`, `syscall`, `time`, `log`).
- Standard library networking used in the pipeline (`net/http`).
- No external storage dependencies are defined in code yet; these will likely be introduced as the pipeline stages evolve.
