package store

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"crawler/internal/shared"
)

// CrawlRecord is a persisted summary of a crawl run.
type CrawlRecord struct {
	ID         string                `json:"id"`
	StartedAt  time.Time             `json:"startedAt"`
	FinishedAt time.Time             `json:"finishedAt"`
	URL        string                `json:"url"`
	Mode       shared.UseCase        `json:"mode"`
	Stats      shared.CrawlStatsView `json:"stats"`
	Error      string                `json:"error,omitempty"`
}

// FileStore appends JSON Lines (one record per line) to a file on disk.
// This keeps persistence simple while still allowing basic querying.
type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

// SaveCrawl appends a crawl record as a single JSON line.
func (s *FileStore) SaveCrawl(_ context.Context, rec CrawlRecord) error {
	if s == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open store file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(&rec); err != nil {
		return fmt.Errorf("encode crawl record: %w", err)
	}

	return nil
}

// ListCrawls reads all records from the JSONL file.
func (s *FileStore) ListCrawls(_ context.Context) ([]CrawlRecord, error) {
	if s == nil {
		return nil, nil
	}

	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open store file: %w", err)
	}
	defer f.Close()

	var out []CrawlRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var rec CrawlRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}
		out = append(out, rec)
	}

	if err := scanner.Err(); err != nil {
		return out, err
	}

	return out, nil
}
