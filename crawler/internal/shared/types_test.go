package shared

import "testing"

func TestSummarizeMode_BlogsReachable(t *testing.T) {
	view := CrawlStatsView{
		TotalRequests: 1,
		Success2xx:    1,
	}

	sum := SummarizeMode(UseCaseTrackBlogs, view)

	if !sum.IsReachable {
		t.Fatalf("expected IsReachable=true, got false")
	}
	if sum.CheckedPages != 1 {
		t.Fatalf("expected CheckedPages=1, got %d", sum.CheckedPages)
	}
	if sum.Message == "" {
		t.Fatalf("expected non-empty message")
	}
}

func TestSummarizeMode_SiteHealth_With5xx(t *testing.T) {
	view := CrawlStatsView{
		TotalRequests:  1,
		ServerError5xx: 1,
	}

	sum := SummarizeMode(UseCaseSiteHealth, view)

	if sum.IsHealthy {
		t.Fatalf("expected IsHealthy=false when 5xx present")
	}
	if sum.Message == "" {
		t.Fatalf("expected non-empty message")
	}
}
