package shared

import (
	"net/http"
	"net/url"
	"sync"
)

type Item struct {
	URL      *url.URL
	Depth    int
	Response *http.Response
}

type WorkTracker struct {
	wg sync.WaitGroup
}

func (w *WorkTracker) Add(n int) {
	w.wg.Add(n)
}

func (w *WorkTracker) Done() {
	w.wg.Done()
}

func (w *WorkTracker) Wait() {
	w.wg.Wait()
}
