### (Temporary but Explicit)

We'll assume:

## 1. Pipeline unit

        We’ll use a single struct flowing through all stages.

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

