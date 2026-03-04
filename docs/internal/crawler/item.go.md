# internal/crawler/item.go

## 1. Overview
- Current state: The source file `crawler/internal/crawler/item.go` is currently empty.
- Where the item type lives now: `crawler/internal/shared/types.go` defines `shared.Item`.
- Why: The crawler + pipeline share a single `Item` type (URL + depth + response + mode + discovered URLs), so it was moved into `internal/shared`.

## 2. File Location
- Relative path (from repo root): `crawler/internal/crawler/item.go`

## 3. Key Components
See `shared.Item` in `crawler/internal/shared/types.go`:

- `URL *url.URL`: Target URL.
- `Depth int`: Crawl depth (seed is depth 0).
- `Response *http.Response`: Set by the fetch stage.
- `Mode shared.UseCase`: Selected use case for the crawl.
- `DiscoveredURLs []string`: Absolute URLs discovered during parse, consumed by the discover stage.

## 4. Execution Flow
- `shared.Item` is a data container and does not define its own control flow.
- It is created by the seeder and discover stage, deduped by the scheduler, populated by fetch/parse stages, then recycled back into scheduling.

## 5. Data Flow
- **Inputs**
	- Parsed `*url.URL` values and associated depth levels from code that constructs `Item` instances.
- **Processing steps**
	- Transformations are performed by other components (scheduler, pipeline); this file only defines the shape of the data.
- **Outputs**
	- `Item` values flowing between internal components.
- **Dependencies**
	- `net/url` and `net/http` from the standard library.
	- Internal: `crawler/internal/shared`.

## 6. Mermaid Diagrams
```mermaid
flowchart LR
	A["Create Item (URL, Depth)"] --> B["Pass through scheduler"]
	B --> C["Pass through pipeline stages"]
```

## 7. Error Handling & Edge Cases
- This file defines a type only; no validation is performed here.
- Callers are responsible for providing non-nil `URL` values and sensible `Depth` values.

## 8. Example Usage
```go
u, _ := url.Parse("https://example.com")
item := shared.Item{URL: u, Depth: 0, Mode: shared.UseCaseTrackBlogs}
// Response is populated by fetch workers in the pipeline.
```

### Notes
#### Why a pointer to `url.URL`?
- Parsing once avoids repeated allocations.
- Ownership is clear.