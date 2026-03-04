package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crawler/internal/service"
	"crawler/internal/shared"
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
}

type startCrawlRequest struct {
	URL     string `json:"url"`
	Workers int    `json:"workers"`
	Depth   int    `json:"depth"`
	Mode    string `json:"mode"`
	Timeout int    `json:"timeoutSeconds"`
}

type startCrawlResponse struct {
	URL   string                `json:"url"`
	Mode  string                `json:"mode"`
	Stats shared.CrawlStatsView `json:"stats"`
	Error string                `json:"error,omitempty"`
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
		URL:   resp.URL,
		Mode:  string(resp.Mode),
		Stats: resp.Stats,
		Error: resp.Err,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
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
