# Data Flow

## Overview
- The project is structured as a concurrent web crawler with pluggable pipeline stages.
- Most internal implementation is not yet present; this document describes the intended data flow suggested by file and package names.

## End-to-End Flow (Conceptual)
- **Inputs**
  - Crawl jobs (e.g., URLs or resources to fetch), provided to the crawler by higher-level code or configuration.
- **Processing**
  - Core crawler receives jobs and passes them through a scheduler.
  - Pipeline stages process each item sequentially: fetch → filter → parse → store.
- **Outputs**
  - Processed items and stored results (exact format and destination are not yet implemented).

## Mermaid Data Flow Diagram
```mermaid
flowchart LR
    A[Input crawl jobs] --> B[Core crawler<br/>(internal/crawler)]
    B --> C[Scheduler<br/>(scheduler.go, planned)]
    C --> D[Fetch stage<br/>(fetch.go, planned)]
    D --> E[Filter stage<br/>(filter.go, planned)]
    E --> F[Parse stage<br/>(parse.go, planned)]
    F --> G[Store stage<br/>(store.go, planned)]
    G --> H[Outputs<br/>(stored results)]
```

## External Dependencies
- Standard library packages used in the entrypoint (`context`, `os/signal`, `syscall`, `time`, `log`).
- No external network or storage dependencies are defined in code yet; these will likely be introduced inside the empty `internal/*` files as the implementation evolves.
