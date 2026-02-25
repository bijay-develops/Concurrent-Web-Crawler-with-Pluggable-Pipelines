package crawler

import (
	"net/http"
	"net/url"
)

type Item struct {
	URL      *url.URL
	Depth    int
	Response *http.Response
}
