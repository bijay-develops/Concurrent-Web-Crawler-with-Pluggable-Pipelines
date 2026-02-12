### This is the brain, not a god object.

``` internal/crawler/crawler.go  ```

### Why this looks “empty”

#### Because control flow comes before work.

### Most junior Go code:
- spawns goroutines first
- figures out shutdown later
- deadlocks under pressure
#### We are doing the opposite.