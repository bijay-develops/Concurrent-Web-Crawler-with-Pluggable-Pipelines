package crawler 

import "net/url"

type Item struct {
	URL *url.URL
	Depth int
}