# Running the Concurrent Web Crawler

This guide assumes you have Go 1.25+ installed.

---

## 1. Project layout

The Go module lives in the `crawler/` subfolder:

```text
Concurrent-Web-Crawler-with-Pluggable-Pipelines/
  crawler/
    cmd/
      crawler/  # CLI entry
      webui/    # Web UI entry
    internal/
      ...
```

Always run Go commands from inside the `crawler/` directory.

---

## 2. Build and sanity check

From the repository root:

```bash
cd crawler

go build ./...
```

If this succeeds, the code compiles correctly.

---

## 3. Running from the CLI

From inside `crawler/`:

```bash
cd crawler

go run ./cmd/crawler \
  -url="https://example.com" \
  -workers=8 \
  -depth=2
```

Flags:

- `-url` – seed URL or domain (e.g. `https://google.com` or `facebook.com`).
  - If you omit the scheme, the crawler assumes `https://`.
- `-workers` – number of concurrent workers (default `8`).
- `-depth` – maximum crawl depth from the seed (default `2`).

Stop the crawler at any time with `Ctrl + C`. The log may show `context canceled` on shutdown; this is a normal, clean exit.

---

## 4. Running the Web UI

From inside `crawler/`:

```bash
cd crawler

go run ./cmd/webui
```

Then open your browser and visit:

- http://localhost:8080/

In the page you can:

- Enter a seed URL or domain (for example `https://google.com` or `chatgpt.com`).
- Set `Workers` and `Max depth`.
- Click **Start crawl**.

The result panel will show:

- `Crawl finished successfully.` for a normal run, or
- An error message if something went wrong (for example, an invalid URL).

---

## 5. Notes and limitations

- The current implementation fetches the seed URL and processes it through the pipeline.
- The `DiscoverWorker` stage is a placeholder; it does not yet extract and enqueue links, so the crawler does not recursively explore the full site.
- Rate limiting is applied per domain via the `DomainLimiter` in the fetch stage.

This manual should be enough to build and run both the CLI and the Web UI for testing and demos.