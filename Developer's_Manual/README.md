# Developer's Manual

This guide explains how to use the concurrent web crawler as a **tool inside your own applications**, especially with stacks like **Next.js + Node.js/Go backends** and **Postgres/MySQL/MongoDB**.

## 1. Run the crawler as a service

The crawler is exposed as a JSON HTTP API.

- Standalone API server:
  - From `crawler/` run: `go run ./cmd/api` (default `:8090`).
- Or via the Web UI server:
  - From `crawler/` run: `go run ./cmd/webui` (default `:8080`, also exposes `/api/crawls` and `/api/crawls/history`).

Key endpoints (HTTP JSON):

- `POST /api/crawls` – start a crawl and get back stats + a human-friendly summary.
- `GET /api/crawls/history` – list past crawl summaries (from `crawler/data/crawls.jsonl`).

As a fullstack dev, treat this as a **crawl microservice** that any app in your ecosystem can call.

## 2. Integrate from Next.js (or any Node.js backend)

Typical pattern:

1. Run the crawler API (e.g. on `http://crawler-api:8090` in Docker, or `http://localhost:8090` in dev).
2. Call `POST /api/crawls` from your Next.js API route / server action.
3. Store or display the returned `stats` and `summary` in your own DB / UI.

### 2.1 Next.js route handler (Node.js)

```ts
// app/api/crawl/route.ts (Next.js App Router example)
import { NextResponse } from 'next/server';

export async function POST(req: Request) {
  const body = await req.json();
  const { url, mode = 'blogs' } = body;

  const res = await fetch(process.env.CRAWLER_URL ?? 'http://localhost:8090/api/crawls', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url, workers: 4, depth: 1, mode, timeoutSeconds: 30 }),
  });

  const data = await res.json();

  // Here you could persist `data` into Postgres/MySQL/MongoDB,
  // or just return it to the frontend.
  return NextResponse.json(data);
}
```

On the frontend, you call your own `/api/crawl` and render `data.summary.message` plus raw `stats`.

### 2.2 Environment configuration

In development:

- Run the crawler API locally and keep `CRAWLER_URL` unset (the default `http://localhost:8090/api/crawls` is used).

In Docker / production:

- Set `CRAWLER_URL` to the internal service URL, for example: `http://crawler-api:8090/api/crawls`.

This keeps your Next.js code the same across environments.

## 3. Persist results in Postgres / MySQL / MongoDB

The crawler already writes JSONL records to `crawler/data/crawls.jsonl`. In your own system, you typically:

- Create a table/collection like `crawl_runs` with fields:
  - `id`, `url`, `mode`, `status_code`, `summary_message`, `stats_json`, timestamps, etc.
- After calling `POST /api/crawls`, insert the response into your DB.

Example fields to store:

- `url` – from `response.url`.
- `mode` – from `response.mode`.
- `status_code` – from `response.stats.lastStatusCode`.
- `summary_message` – from `response.summary.message`.
- `raw_stats` – JSON column / document with the whole `response.stats`.

That gives you:

- Dashboards for site health over time.
- Ability to query for failing URLs.
- A history of crawls you can join with other business data.

### 3.1 Postgres / MySQL schema example

For relational databases, you can use a table like:

```sql
CREATE TABLE crawl_runs (
  id            bigserial PRIMARY KEY,
  url           text NOT NULL,
  mode          text NOT NULL,
  status_code   integer,
  summary       text,
  stats_json    jsonb,
  created_at    timestamptz DEFAULT now()
);
```

Insert from Node.js (pseudo-code):

```ts
await db.query(
  'INSERT INTO crawl_runs (url, mode, status_code, summary, stats_json) VALUES ($1,$2,$3,$4,$5)',
  [
    data.url,
    data.mode,
    data.stats.lastStatusCode,
    data.summary.message,
    data.stats,
  ],
);
```

For MySQL, use `JSON` instead of `jsonb` and adapt the types as needed.

### 3.2 MongoDB collection example

For MongoDB, create a `crawl_runs` collection and insert documents like:

```ts
await db.collection('crawl_runs').insertOne({
  url: data.url,
  mode: data.mode,
  statusCode: data.stats.lastStatusCode,
  summary: data.summary.message,
  stats: data.stats,
  createdAt: new Date(),
});
```

From there you can build dashboards, alerts, or analytics over this collection.

## 4. Use it for automated site-health checks (CI/CD)

You can run this project as part of your CI pipeline:

1. Build and run the API container (see `crawler/Dockerfile.api` and `docker-compose.yml`).
2. In a CI job, call `POST /api/crawls` against your staging/production URL.
3. Parse the JSON and fail the job if, for example:
   - `summary.isHealthy` is `false`, or
   - `stats.ServerError5xx` > 0.

This gives you a very simple, programmable **smoke test** for your deployed app.

## 5. Track blogs or content feeds

With `mode = blogs`:

- Schedule periodic crawls of your own blogs or external feeds you care about.
- Store the results in your DB.
- Alert if a blog becomes unreachable or starts returning 4xx/5xx.

This can be part of:

- A personal dashboard.
- A content-operations tool your team uses.

## 6. Feed search / analytics pipelines

With `mode = search`:

- Use this service as the **front door** to a richer pipeline:
  - The crawler tells you which URLs are reachable and suitable for indexing (`summary.isIndexable`, `stats.Success2xx`).
  - A separate job fetches full HTML/content and pushes documents into Elasticsearch / OpenSearch / Meilisearch / your analytics DB.

This separation keeps the **web-facing concurrent crawling** logic in one place, and your **domain-specific indexing** logic in your own services.

## 7. Learn and reuse concurrency patterns (Go backend)

Internally, the crawler shows:

- Worker pools and channel-based pipelines (`internal/crawler`, `internal/pipeline`).
- Per-domain rate limiting (`internal/pipeline/limiter.go`).
- Backpressure via `WorkTracker`.

When you build other Go microservices, you can:

- Reuse these patterns (or copy the relevant types) as a starting point.
- Experiment here first, then port to production code.

## 8. Portfolio, demos, and teaching

Because this repo has:

- A Web UI.
- A JSON API.
- A service layer and persistence.
- CI and Docker files.

…you can also use it to:

- Demonstrate back-end + front-end integration in interviews or talks.
- Show junior developers an end-to-end example of a small but realistic distributed tool.

---

In short, run the crawler as a **separate service**, call it from your **Next.js/Node/Go backends**, and store the JSON responses in your own databases. Use the three modes (blogs / health / search) to tailor how you interpret and act on the results in your products.
