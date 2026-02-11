### Why a pointer to url.URL?
- Parsing once avoids repeated allocations
- Ownership is clear