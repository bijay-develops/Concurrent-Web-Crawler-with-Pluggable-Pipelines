# Fix 1: go.mod not found when running `go mod tidy`

## Problem

From the project root, running:

```bash
go mod tidy
```

produced:

```text
go: go.mod file not found in current directory or any parent directory; see 'go help modules'
```

This was confusing because the repository *does* contain a `go.mod` file under `crawler/go.mod`.

## Root Cause

Go modules are discovered starting from the **current directory** and walking **upward**. When you ran `go mod tidy` in the repository root (`Concurrent-Web-Crawler-with-Pluggable-Pipelines/`), there was no `go.mod` in that directory or any parent directory, so Go treated it as a non‑module directory and failed.

Our actual module is defined in the subfolder:

```text
crawler/go.mod
```

with:

```go
module crawler

go 1.25
```

So any module-aware commands (`go run`, `go test`, `go build`, `go mod tidy`, etc.) must be executed **inside** the `crawler/` folder (or a subfolder of it).

## Fix

1. Change directory into the module root:

   ```bash
   cd crawler
   ```

2. Run `go mod tidy` again:

   ```bash
   go mod tidy
   ```

3. This time it succeeds because Go finds `crawler/go.mod` in the current directory.

## Key Takeaways

- Go module commands operate relative to the **nearest `go.mod` above the current directory**.
- If you see `go.mod file not found`, verify that you are:
  - In the correct subdirectory, and
  - That `go.mod` is actually present there.
- In this project, always run Go tooling from the `crawler/` directory (or below), not from the repository root.