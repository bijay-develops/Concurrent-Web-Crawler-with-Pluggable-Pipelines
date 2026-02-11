# Architecture

## Overview
- Entry point: Go CLI in `crawler/cmd/crawler/main.go`.
- Core crawler logic and public API planned under `crawler/internal/crawler/`.
- Pluggable pipeline stages planned under `crawler/internal/pipeline/`.
- Most implementation files are currently empty placeholders; this document reflects the intended structure implied by package and file names.

## Modules
- `cmd/crawler/main.go`
  - Sets up process context and signal handling.
  - Constructs a crawler configuration and starts the crawler.
- `internal/crawler/`
  - Intended home for the main crawler type, configuration, and orchestration logic.
  - Likely to expose `Config`, `New`, and `Run` as referenced from `main.go`.
- `internal/pipeline/`
  - Intended pluggable pipeline stages (`fetch`, `filter`, `parse`, `store`).
  - Each stage is separated into its own file for clarity and independent evolution.

## High-Level Interactions
- The CLI entrypoint configures and starts the crawler through the internal package API.
- The core crawler is expected to coordinate scheduling of crawl work and interaction with pipeline stages.
- Pipeline stages will process crawl items in sequence (e.g., fetch → filter → parse → store).

## Mermaid System Diagram
```mermaid
flowchart TD
    CLI[cmd/crawler/main.go<br/>CLI entrypoint] --> CORE[internal/crawler<br/>Core crawler]
    CORE --> SCHED[internal/crawler/scheduler.go<br/>Scheduler (planned)]
    CORE --> PIPE[internal/pipeline<br/>Pipeline (planned)]
    PIPE --> FETCH[fetch.go<br/>Fetch stage (planned)]
    PIPE --> FILTER[filter.go<br/>Filter stage (planned)]
    PIPE --> PARSE[parse.go<br/>Parse stage (planned)]
    PIPE --> STORE[store.go<br/>Store stage (planned)]
```

## Notes
- The above diagram represents the intended architecture based on the current package layout and function calls in `main.go`.
- As of now, the `internal/*` files are empty; implementation details will refine this architecture in future revisions.
