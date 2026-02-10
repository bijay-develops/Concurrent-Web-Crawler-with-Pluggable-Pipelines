``` 
crawler/
├── cmd/
│   └── crawler/
│       └── main.go
├── internal/
│   ├── crawler/
│   │   ├── crawler.go
│   │   ├── scheduler.go
│   │   └── item.go
│   └── pipeline/
│       ├── fetch.go
│       ├── parse.go
│       ├── filter.go
│       └── store.go
└── go.mod

```
---
<br>

Why this matters: 
- cmd/ keeps binaries thin
- internal/ enforces encapsulation
- Pipeline stages are visible but controlled