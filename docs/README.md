# Project Documentation

This `docs/` directory contains structured documentation for the crawler project. The documentation mirrors the source tree under `crawler/`, excluding the top-level `crawler` directory itself.

At a high level, the repository layout is:

```text
Concurrent-Web-Crawler-with-Pluggable-Pipelines/
  crawler/         # Go module (CLI + Web UI + internal packages)
    cmd/
      crawler/     # CLI entrypoint
      webui/       # Web UI entrypoint
    internal/
      crawler/     # Core crawler orchestration
      pipeline/    # Pluggable pipeline stages and rate limiting
      shared/      # Shared types such as Item and UseCase
  docs/            # This documentation tree
  manual/          # How to build and run the project
  WHY_the_PROJECT/ # High-level motivation and use-case explanations
  Problems-and-Solutions/
  QnA/
  code_fixing/
```

## Index

### Project-Level Docs

- [ARCHITECTURE](ARCHITECTURE.md) – overall system structure and module interactions.
- [DATA_FLOW](DATA_FLOW.md) – conceptual end-to-end data flow and pipeline stages.

### Source File Docs

- [go.mod](go.mod.md)
- [cmd/crawler/main.go](cmd/crawler/main.go.md)
- [internal/crawler/crawler.go](internal/crawler/crawler.go.md)
- [internal/crawler/item.go](internal/crawler/item.go.md)
- [internal/crawler/scheduler.go](internal/crawler/scheduler.go.md)
- [internal/crawler/work.go](internal/crawler/work.go.md)
- [internal/pipeline/fetch.go](internal/pipeline/fetch.go.md)
- [internal/pipeline/filter.go](internal/pipeline/filter.go.md)
- [internal/pipeline/parse.go](internal/pipeline/parse.go.md)
- [internal/pipeline/store.go](internal/pipeline/store.go.md)
- [internal/pipeline/discover.go](internal/pipeline/discover.go.md)
- [internal/pipeline/interfaces.go](internal/pipeline/interfaces.go.md)
- [internal/pipeline/limiter.go](internal/pipeline/limiter.go.md)

## How to Navigate

- Use the links above to jump to documentation for a specific source file.
- Each file-level document follows a consistent structure:
  1. Overview
  2. File Location
  3. Key Components
  4. Execution Flow
  5. Data Flow
  6. Mermaid Diagrams
  7. Error Handling & Edge Cases
  8. Example Usage
- Where a source file is currently empty, the documentation explicitly notes that and only describes the intended role implied by its name and placement.
