# internal/pipeline/fetch.go

## 1. Overview
- Purpose: Intended to implement the "fetch" stage of the pipeline that performs HTTP requests for crawl items.
- Current state: The file exists in `internal/pipeline` but is empty; this document describes the planned role.
- High-level responsibility (implied): Take scheduled `Item` values, issue HTTP requests, and attach responses.

## 2. File Location
- Relative path (from repo root): `crawler/internal/pipeline/fetch.go`

## 3. Key Components (Planned)
- Worker functions that:
  - Read `Item` values from an input channel.
  - Use a shared HTTP client (possibly with rate limiting) to perform requests.
  - Populate the `Response` field on each `Item`.
  - Forward results to the next stage (e.g., parse).

## 4. Execution Flow (Planned)
1. The core crawler or scheduler enqueues `Item` values onto a "scheduled" channel.
2. One or more fetch workers read from that channel.
3. For each `Item`, a worker performs an HTTP request and attaches the `*http.Response`.
4. The enriched `Item` is written to a "fetched" channel for downstream stages.

## 5. Data Flow (Planned)
- **Inputs**
  - `Item` values with populated `URL` and `Depth`.
- **Processing steps**
  - Perform HTTP GET (or other) requests.
  - Handle transient errors, timeouts, and retries.
- **Outputs**
  - `Item` values with a populated `Response` field.
- **Dependencies**
  - Standard library HTTP client and any shared limiter from `internal/pipeline/limiter.go`.

## 6. Mermaid Diagrams (Conceptual)
```mermaid
flowchart LR
  A["Scheduled items"] --> B["Fetch workers (planned)"]
  B --> C["Fetched items with Response"]
```

## 7. Error Handling & Edge Cases (Planned)
- Network errors, timeouts, and non-2xx status codes must be handled gracefully.
- Resources like response bodies must be closed to avoid leaks.
- Backoff and retry strategies may be implemented where appropriate.

## 8. Example Usage
- No concrete API exists yet; once implemented, this stage will be invoked from `internal/crawler/crawler.go` via helper functions in `internal/pipeline`.
