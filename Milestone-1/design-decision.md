### (Temporary but Explicit)

### We'll assume:

## 1. Pipeline unit
#### We’ll use a single struct flowing through all stages.
### Why?
- Ownership is explicit
- Easy to add/remove fields
- Easy to reason about memory later

```bash
type Item struct {
	URL   *url.URL
	Depth int
}
```
#### No body, no response yet. We add fields only when needed.


## 2. Deduplication
#### Before fetch, URL-level only.
### Why?
- Cheapest place to drop work
- Avoids waiting network + memory
- Content-hash dedupe can be added later

#### This means we need a concurrent-safe visited set.