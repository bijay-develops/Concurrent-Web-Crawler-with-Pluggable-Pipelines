# Fix 3: Transient compile errors after refactor

After introducing the `internal/shared` package and moving `Item` and `WorkTracker` there, we saw several temporary compile errors. These are useful to document because they are common when refactoring Go code.

---

## 1. `missing metadata for import` (before refactor)

Earlier, before the import-cycle fix, we saw:

```text
could not import crawler/internal/pipeline (missing metadata for import of "crawler/internal/pipeline")
```

### What it meant

This usually indicates that Go could not build or locate the imported package's compiled metadata. Common causes are:

- The imported package does not compile.
- `go list` / `go env` cannot traverse the module correctly (e.g., wrong working directory).

In our case, it was a side-effect of:

- Running tooling from the wrong directory, and
- The import cycle preventing `crawler/internal/pipeline` from compiling cleanly.

### How it was resolved

- Running `go mod tidy` from the correct module root (`crawler/`).
- Breaking the import cycle via the `internal/shared` package (see `02-import-cycle-crawler-pipeline.md`).

Once the package graph was valid and the module root was correct, this error disappeared.

---

## 2. `undefined: shared` and `undefined: pipeline`

After we created `internal/shared` and started using `shared.Item` and `shared.WorkTracker`, we initially forgot to import the new package in some files. For example, in `internal/crawler/crawler.go` we saw:

```text
undefined: shared
```

and in the same file, after another edit:

```text
undefined: pipeline
```

### What it meant

- `undefined: shared` – the code referenced the identifier `shared` (e.g. `shared.Item`) but there was **no import** of `crawler/internal/shared` in that file.
- `undefined: pipeline` – similarly, calls like `pipeline.FetchWorker` existed, but the `crawler/internal/pipeline` import had been temporarily removed or not yet re-added.

### How we fixed it

For each affected file, we:

- Ensured the correct imports were present, e.g. in `internal/crawler/crawler.go`:
  - Import `crawler/internal/shared`.
  - Import `crawler/internal/pipeline`.
- Replaced any lingering references to old types with the new ones:
  - Channels and values now use `shared.Item`.
  - Work tracking uses `*shared.WorkTracker`.

Once imports matched the identifiers used in the code, these `undefined` errors went away.

---

## 3. `imported and not used` and `undefined: sync`

During cleanup, we also saw typical Go linter/compiler messages such as:

```text
"crawler/internal/shared" imported and not used
```

and in another context:

```text
undefined: sync
```

### What they meant

- `imported and not used` – Go does not allow unused imports. If you import a package but do not reference anything from it in that file, the compiler fails.
- `undefined: sync` – we were still using `sync.WaitGroup` in `internal/crawler/work.go` *after* moving the type to `internal/shared/types.go`, but the `sync` package was not imported (or we intended to stop using it directly in that file).

### How we fixed them

- For **unused imports**:
  - If a package truly was no longer needed (e.g., `crawler/internal/shared` in a file that no longer used `shared.Item`), we removed the import line.
- For **`sync` usage** in `work.go`:
  - We decided that `WorkTracker` would live in `internal/shared` and that `work.go` itself no longer needed to declare the struct.
  - We cleaned up imports so that the file either:
    - Uses `shared.WorkTracker` from `internal/shared`, or
    - If the file only contained the wrapper logic, we kept the `sync` import and ensured it matched the actual code.

The key idea is that **every imported package must be used, and every external identifier must have a matching import**.

---

## Lessons Learned

- Large refactors often create a *wave* of small, local compile errors. This is normal.
- A good process is:
  1. Make a structural change (like introducing `internal/shared`).
  2. Fix import paths and types file by file.
  3. Re-run `go build ./...` or your editor's analyzer, and address remaining errors from top to bottom.
- Most Go compiler messages are very literal:
  - `imported and not used` → remove or use the import.
  - `undefined: X` → either add the correct import or rename the identifier.
  - `import cycle not allowed` → requires a design-level fix to your package graph.

By following this approach, we arrived at a clean, compiling project with a clearer separation of concerns between `crawler`, `pipeline`, and `shared`.