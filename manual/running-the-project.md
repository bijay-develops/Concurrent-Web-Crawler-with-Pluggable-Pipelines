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
  -depth=2 \
  -mode=1
```

Flags:

- `-url` – seed URL or domain (e.g. `https://google.com` or `facebook.com`).
  - If you omit the scheme, the crawler assumes `https://`.
  - The examples here sometimes use `https://example.com`, which is a standard documentation domain. On some machines with unusual TLS settings, this may fail trust checks; if you see a certificate error, simply try another public site you control.
- `-workers` – number of concurrent workers (default `8`).
- `-depth` – maximum crawl depth from the seed (default `2`).
 - `-mode` – high-level use case. You can pass:
   - `1` or `blogs` – Track my favourite blogs.
   - `2` or `health` – Internal Site Health Checker.
   - `3` or `search` – Data Pipeline Search Index.

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
- Choose one of three **Use case** options:
  1. Track my favourite blogs
  2. Internal Site Health Checker
  3. Data Pipeline Search Index
- Click **Start crawl**.

The result panel will show:

- `Crawl finished successfully.` for a normal run, or
- An error message if something went wrong (for example, an invalid URL).
- A "What this means" section that interprets the stats for the selected mode (blogs / site health / search index).

---
## 5. Running the JSON API server

From inside `crawler/`:

```bash
cd crawler

go run ./cmd/api
```

By default the API listens on port `8090`. You can override this with:

```bash
API_PORT=9090 go run ./cmd/api
```

Key endpoints:

- `POST /api/crawls`
  - JSON body (all fields optional except `url`):
    - `url`: seed URL to crawl.
    - `workers`: worker count.
    - `depth`: maximum depth.
    - `mode`: `blogs` | `health` | `search` (or their numeric aliases).
    - `timeoutSeconds`: optional timeout.
  - Returns JSON with fields like `url`, `mode`, `stats`, `summary`, and `error`.

- `GET /api/crawls/history`
  - Returns an array of past crawl summaries loaded from `data/crawls.jsonl`.

Example `curl` to start a crawl:

```bash
curl -X POST http://localhost:8090/api/crawls \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com","workers":4,"depth":1,"mode":"blogs"}'
```

---

## 6. Running with Docker and docker-compose

From the repository root:

```bash
docker compose up --build webui
```

This will:

- Build the Web UI container from `crawler/Dockerfile.webui`.
- Expose the Web UI on `http://localhost:8080` (internally configurable via `WEBUI_PORT`).

To run the standalone API container:

```bash
docker compose up --build api
```

This will:

- Build the API container from `crawler/Dockerfile.api`.
- Expose the API on `http://localhost:8090` (internally configurable via `API_PORT`).

---

## 7. Notes and limitations

- The crawler fetches the seed URL and can **recursively explore internal links** up to the configured `-depth`.
  - Discovery is conservative: it prefers post-like permalinks and avoids common noisy internal areas (tag/category/author/archive links except safe pagination).
  - On listing pages (many `<article>` cards), discovery prefers the “main permalink per card” to keep results cleaner.
- A global **max unique URLs per crawl** cap is enforced by the scheduler (default `40` in the core crawler). This is intentional to keep demos safe.
- Rate limiting is applied per domain via `DomainLimiter` in the fetch stage.
- The parser only attempts HTML analytics for `text/html` / `application/xhtml+xml` and reads a **bounded** amount of HTML per page (currently capped to 4 MiB) to avoid memory spikes.
- If the server responds with `429 Too Many Requests`, fetch workers briefly back off.
- Crawl history is appended as JSON Lines to `crawler/data/crawls.jsonl` and surfaced via the API and Web UI.
  - History reading supports larger records (scanner buffer raised above the default 64 KiB).
- TLS behavior depends on your local system trust store. If `https://example.com` or another HTTPS URL fails with a certificate error, that indicates an environment trust issue, not a problem with the crawler itself.

This manual should be enough to build and run the CLI, Web UI, and JSON API directly or via Docker for testing and demos.