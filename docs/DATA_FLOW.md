# Data Flow

## Overview
- The project is structured as a concurrent web crawler with pluggable pipeline stages.
- Core control flow and scheduling are implemented; pipeline stages are present as structural placeholders that will be wired in.

## End-to-End Flow (Conceptual)
- **Inputs**
  - Crawl jobs (e.g., URLs or resources to fetch), provided to the crawler by higher-level code or configuration.
- **Processing**
  - Core crawler receives jobs and passes them through a scheduler that deduplicates URLs.
  - Pipeline stages process each item conceptually in sequence: discover → fetch → filter → parse → store, with limiting where appropriate.
- **Outputs**
  - Processed items and stored results (exact format and destination are not yet implemented).

## Mermaid Data Flow Diagram
```mermaid
flowchart LR
  A[Input crawl jobs] --> B[Core crawler<br/>(internal/crawler)]
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
- No external network or storage dependencies are defined in code yet; these will likely be introduced inside the empty `internal/*` files as the implementation evolves.
