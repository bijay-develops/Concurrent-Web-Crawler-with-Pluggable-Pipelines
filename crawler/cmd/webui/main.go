package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"crawler/internal/crawler"
	"crawler/internal/httpapi"
	"crawler/internal/shared"
)

var pageTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Concurrent Web Crawler</title>
  <style>
    body { font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; padding: 2rem; background: #0f172a; color: #e5e7eb; }
    h1 { margin-bottom: 0.5rem; }
    p { color: #9ca3af; }
    form { margin-top: 1.5rem; padding: 1.5rem; background: #020617; border-radius: 0.75rem; border: 1px solid #1e293b; max-width: 480px; }
    label { display: block; margin-top: 1rem; font-size: 0.9rem; color: #9ca3af; }
    input[type="text"], input[type="number"] { width: 100%; padding: 0.5rem 0.75rem; margin-top: 0.25rem; border-radius: 0.5rem; border: 1px solid #1f2937; background: #020617; color: #e5e7eb; }
    input[type="text"]:focus, input[type="number"]:focus { outline: none; border-color: #38bdf8; box-shadow: 0 0 0 1px #38bdf8; }
    button { margin-top: 1.5rem; padding: 0.6rem 1.2rem; border-radius: 999px; border: none; background: linear-gradient(to right, #0ea5e9, #22c55e); color: #020617; font-weight: 600; cursor: pointer; }
    button:hover { filter: brightness(1.05); }
    .result { margin-top: 1.5rem; padding: 1rem 1.25rem; border-radius: 0.75rem; border: 1px solid #1e293b; background: #020617; max-width: 480px; font-size: 0.9rem; }
    .error { color: #fecaca; }
    .ok { color: #bbf7d0; }
    code { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; font-size: 0.85rem; }
		.mode-group { margin-top: 1rem; }
		.mode-title { display: block; font-size: 0.9rem; color: #9ca3af; margin-bottom: 0.4rem; }
		.mode-option { display: block; margin-top: 0.25rem; font-size: 0.85rem; }
		.mode-option input { margin-right: 0.35rem; }
  </style>
</head>
<body>
  <h1>Concurrent Web Crawler</h1>
  <p>Run a crawl with custom URL, worker count, and depth.</p>

  <form method="POST" action="/crawl">
    <label>
      Seed URL or domain
      <input type="text" name="url" value="{{.URL}}" placeholder="e.g. https://example.com or google.com" required />
    </label>

    <label>
      Workers
      <input type="number" name="workers" min="1" max="128" value="{{.Workers}}" />
    </label>

    <label>
      Max depth
      <input type="number" name="depth" min="0" max="8" value="{{.Depth}}" />
    </label>

		<div class="mode-group">
			<span class="mode-title">Use case</span>
			<label class="mode-option">
				<input type="radio" name="mode" value="blogs" {{if eq .Mode "blogs"}}checked{{end}} />
				1. Track my favourite blogs
			</label>
			<label class="mode-option">
				<input type="radio" name="mode" value="health" {{if eq .Mode "health"}}checked{{end}} />
				2. Internal Site Health Checker
			</label>
			<label class="mode-option">
				<input type="radio" name="mode" value="search" {{if eq .Mode "search"}}checked{{end}} />
				3. Data Pipeline Search Index
			</label>
		</div>

    <button type="submit">Start crawl</button>
  </form>

	<div class="result">
		{{if .Ran}}
			<div><strong>Result</strong></div>
			{{if .Error}}
				<p class="error">Error: {{.Error}}</p>
			{{else}}
				<p class="ok">Crawl finished successfully.</p>
			{{end}}
			<p><strong>Config</strong></p>
			<p><code>url={{.URL}} workers={{.Workers}} depth={{.Depth}} mode={{.Mode}}</code></p>
			{{if .Summary}}
				<p><strong>What this means</strong></p>
				<p>{{.Summary}}</p>
			{{end}}
		{{else}}
			<div><strong>Result</strong></div>
			<p>Fill the form and start a crawl to see a summary here.</p>
		{{end}}
	</div>
	<script>
		(function() {
			const form = document.querySelector('form');
			if (!form) return;

			form.addEventListener('submit', async function (e) {
				e.preventDefault();

				const formData = new FormData(form);
				const url = formData.get('url');
				const workers = parseInt(formData.get('workers') || '8', 10);
				const depth = parseInt(formData.get('depth') || '2', 10);
				const mode = formData.get('mode') || 'blogs';

				let resultBox = document.querySelector('.result');
				if (!resultBox) {
					resultBox = document.createElement('div');
					resultBox.className = 'result';
					form.insertAdjacentElement('afterend', resultBox);
				}
				resultBox.innerHTML = '<div><strong>Result</strong></div><p>Running crawl...</p>';

				try {
					const res = await fetch('/api/crawls', {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ url, workers, depth, mode, timeoutSeconds: 30 })
					});
					const data = await res.json();

					const stats = data.stats || {};
					const summary = data.summary || {};

					let html = '<div><strong>Result</strong></div>';
					if (data.error) {
						html += '<p class="error">Error: ' + data.error + '</p>';
					} else {
						html += '<p class="ok">Crawl finished successfully.</p>';
					}
					html += '<p><strong>Config</strong></p>';
					html += '<p><code>url=' + (data.url || url) + ' workers=' + workers + ' depth=' + depth + ' mode=' + (data.mode || mode) + '</code></p>';

					if (summary && summary.message) {
						html += '<p><strong>What this means</strong></p>';
						html += '<p>' + summary.message + '</p>';
					}

					// Detailed numeric stats from the crawl
					html += '<p><strong>Detailed stats</strong></p>';
					html += '<ul>';
					html += '<li>Total requests: ' + (stats.totalRequests || 0) + '</li>';
					html += '<li>2xx successes: ' + (stats.success2xx || 0) + '</li>';
					html += '<li>4xx client errors: ' + (stats.clientError4xx || 0) + '</li>';
					html += '<li>5xx server errors: ' + (stats.serverError5xx || 0) + '</li>';
					html += '<li>Other status codes: ' + (stats.otherStatus || 0) + '</li>';
					html += '<li>Network errors: ' + (stats.networkErrors || 0) + '</li>';
					if (stats.lastStatusCode) {
						html += '<li>Last status code: ' + stats.lastStatusCode + '</li>';
					}
					if (stats.lastUrl) {
						html += '<li>Last URL: ' + stats.lastUrl + '</li>';
					}
					if (stats.parsedPages) {
						html += '<li>Pages parsed for content: ' + stats.parsedPages + '</li>';
					}
					if (stats.totalWords) {
						html += '<li>Total words (approx): ' + stats.totalWords + '</li>';
					}
					if (stats.totalInternalLinks || stats.totalExternalLinks) {
						html += '<li>Internal links (sum): ' + (stats.totalInternalLinks || 0) + '</li>';
						html += '<li>External links (sum): ' + (stats.totalExternalLinks || 0) + '</li>';
					}
					html += '</ul>';

					// Use-case specific fields (options 1, 2, 3)
					html += '<p><strong>Use-case details</strong></p>';
					html += '<ul>';
					const modeValue = summary.mode || data.mode || mode;
					html += '<li>Mode: ' + modeValue + '</li>';
					if (typeof summary.checkedPages === 'number') {
						html += '<li>Pages checked: ' + summary.checkedPages + '</li>';
					}
					if (typeof summary.isReachable === 'boolean') {
						html += '<li>Reachable (blogs use case): ' + (summary.isReachable ? 'yes' : 'no') + '</li>';
					}
					if (typeof summary.isHealthy === 'boolean') {
						html += '<li>Site healthy (health checker): ' + (summary.isHealthy ? 'yes' : 'no') + '</li>';
					}
					if (typeof summary.isIndexable === 'boolean') {
						html += '<li>Good for indexing (search index): ' + (summary.isIndexable ? 'yes' : 'no') + '</li>';
					}
					if (summary.primaryStatus) {
						html += '<li>Primary status code: ' + summary.primaryStatus + '</li>';
					}
					if (typeof summary.averageWordsPerPage === 'number' && summary.averageWordsPerPage > 0) {
						html += '<li>Average words per page: ' + summary.averageWordsPerPage + '</li>';
					}
					if (summary.longestPageTitle || summary.longestPageUrl) {
						html += '<li>Largest page: ' +
						  (summary.longestPageTitle ? ('"' + summary.longestPageTitle + '" ') : '') +
						  (summary.longestPageWordCount ? ('(' + summary.longestPageWordCount + ' words) ') : '') +
						  (summary.longestPageUrl ? ('@ ' + summary.longestPageUrl) : '') +
						  '</li>';
					}
					html += '</ul>';

					resultBox.innerHTML = html;
				} catch (err) {
					if (resultBox) {
						resultBox.innerHTML = '<div><strong>Result</strong></div><p class="error">Failed to call API: ' + err + '</p>';
					}
				}
			});
		})();
	</script>
</body>
</html>`))

type pageData struct {
	URL     string
	Workers int
	Depth   int
	Mode    string
	Ran     bool
	Error   string
	Summary string
}

func normalizeMode(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "blogs"
	}
	switch s {
	case "1", "blogs", "blog", "track-blogs", "track my favourite blogs":
		return "blogs"
	case "2", "health", "site-health", "internal site health checker":
		return "health"
	case "3", "search", "index", "search-index", "data pipeline search index":
		return "search"
	default:
		return "blogs"
	}
}

func parseUseCase(mode string) shared.UseCase {
	switch mode {
	case "blogs":
		return shared.UseCaseTrackBlogs
	case "health":
		return shared.UseCaseSiteHealth
	case "search":
		return shared.UseCaseSearchIndex
	default:
		return shared.UseCaseTrackBlogs
	}
}

func buildSummary(useCase shared.UseCase, s shared.CrawlStatsView) string {
	return shared.SummarizeMode(useCase, s).Message
}

func main() {
	mux := http.NewServeMux()

	// Attach JSON API on the same server so the Web UI can
	// talk to it without CORS issues.
	apiHandler := httpapi.NewHandler()
	apiHandler.Register(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := pageData{URL: "https://example.com", Workers: 8, Depth: 2, Mode: "blogs"}
		_ = pageTmpl.Execute(w, data)
	})

	mux.HandleFunc("/crawl", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = pageTmpl.Execute(w, pageData{URL: "", Workers: 8, Depth: 2, Ran: true, Error: "invalid form"})
			return
		}

		urlVal := r.Form.Get("url")
		workersVal := r.Form.Get("workers")
		depthVal := r.Form.Get("depth")
		modeVal := r.Form.Get("mode")

		workers := 8
		if workersVal != "" {
			if n, err := strconv.Atoi(workersVal); err == nil && n > 0 {
				workers = n
			}
		}

		depth := 2
		if depthVal != "" {
			if d, err := strconv.Atoi(depthVal); err == nil && d >= 0 {
				depth = d
			}
		}

		modeStr := normalizeMode(modeVal)
		useCase := parseUseCase(modeStr)

		stats := &shared.CrawlStats{}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		c := crawler.New(
			crawler.WithWorkerCount(workers),
			crawler.WithMaxDepth(depth),
			crawler.WithSeedURL(urlVal),
			crawler.WithUseCase(useCase),
			crawler.WithStatsCollector(stats),
		)

		var runErr string
		if err := c.Run(ctx); err != nil {
			// Treat normal shutdown (context cancelled or timed out) as success
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				runErr = err.Error()
			}
		}

		statsSnapshot := stats.Snapshot()
		summary := buildSummary(useCase, statsSnapshot)

		data := pageData{
			URL:     urlVal,
			Workers: workers,
			Depth:   depth,
			Mode:    modeStr,
			Ran:     true,
			Error:   runErr,
			Summary: summary,
		}
		_ = pageTmpl.Execute(w, data)
	})

	addr := ":8080"
	if v := os.Getenv("WEBUI_PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("Web UI listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
