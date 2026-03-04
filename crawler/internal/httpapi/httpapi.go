package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crawler/internal/service"
	"crawler/internal/shared"
	"crawler/internal/store"
)

// Handler wires HTTP requests to the CrawlService.
type Handler struct {
	crawls *service.CrawlService
}

func NewHandler() *Handler {
	return &Handler{crawls: service.NewCrawlService()}
}

// Register attaches the API routes to a mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/crawls", h.handleStartCrawl)
	mux.HandleFunc("/api/crawls/history", h.handleListCrawls)
}

type startCrawlRequest struct {
	URL     string `json:"url"`
	Workers int    `json:"workers"`
	Depth   int    `json:"depth"`
	Mode    string `json:"mode"`
	Timeout int    `json:"timeoutSeconds"`
}

type startCrawlResponse struct {
	URL     string                `json:"url"`
	Mode    string                `json:"mode"`
	Stats   shared.CrawlStatsView `json:"stats"`
	Summary shared.ModeSummary    `json:"summary"`
	Error   string                `json:"error,omitempty"`
}

type crawlHistoryResponse struct {
	Crawls []map[string]any `json:"crawls"`
}

func (h *Handler) handleStartCrawl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req startCrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	mode := parseMode(req.Mode)

	// Allow overriding from query params for quick testing
	if q := r.URL.Query().Get("workers"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			req.Workers = n
		}
	}

	if q := r.URL.Query().Get("depth"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			req.Depth = n
		}
	}

	resp, _ := h.crawls.StartCrawl(r.Context(), service.StartRequest{
		URL:     req.URL,
		Workers: req.Workers,
		Depth:   req.Depth,
		Mode:    mode,
		Timeout: time.Duration(req.Timeout) * time.Second,
	})

	out := startCrawlResponse{
		URL:     resp.URL,
		Mode:    string(resp.Mode),
		Stats:   resp.Stats,
		Summary: resp.Summary,
		Error:   resp.Err,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handler) handleListCrawls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// For now, reach into the FileStore via the concrete type.
	// If the service store is nil, just return an empty list.
	fs := h.crawlsStore()
	if fs == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(crawlHistoryResponse{Crawls: nil})
		return
	}

	recs, err := fs.ListCrawls(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	out := make([]map[string]any, 0, len(recs))
	for _, rec := range recs {
		out = append(out, map[string]any{
			"id":         rec.ID,
			"startedAt":  rec.StartedAt,
			"finishedAt": rec.FinishedAt,
			"url":        rec.URL,
			"mode":       rec.Mode,
			"stats":      rec.Stats,
			"summary":    shared.SummarizeMode(rec.Mode, rec.Stats),
			"error":      rec.Error,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(crawlHistoryResponse{Crawls: out})
}

// crawlsStore is a helper to access the underlying FileStore when present.
func (h *Handler) crawlsStore() interface {
	ListCrawls(context.Context) ([]store.CrawlRecord, error)
} {
	// At the moment, CrawlService always uses a FileStore with a known path,
	// so we can construct a matching store here. In future this could be
	// injected instead of re-created.
	return store.NewFileStore("data/crawls.jsonl")
}

func parseMode(s string) shared.UseCase {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "1", "blogs", "blog", "track-blogs", "track my favourite blogs":
		return shared.UseCaseTrackBlogs
	case "2", "health", "site-health", "internal site health checker":
		return shared.UseCaseSiteHealth
	case "3", "search", "index", "search-index", "data pipeline search index":
		return shared.UseCaseSearchIndex
	default:
		return shared.UseCaseTrackBlogs
	}
}
