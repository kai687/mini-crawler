package recordutil

import "testing"

func TestURLWithAnchor(t *testing.T) {
	got := URLWithAnchor("https://example.com/page", "section")

	want := "https://example.com/page#section"
	if got != want {
		t.Fatalf("URLWithAnchor() = %q, want %q", got, want)
	}
}

func TestURLWithAnchorEmptyAnchor(t *testing.T) {
	got := URLWithAnchor("https://example.com/page", "")

	want := "https://example.com/page"
	if got != want {
		t.Fatalf("URLWithAnchor() = %q, want %q", got, want)
	}
}

func TestURLWithAnchorReplacesExistingAnchor(t *testing.T) {
	got := URLWithAnchor("https://example.com/page#old", "section")

	want := "https://example.com/page#section"
	if got != want {
		t.Fatalf("URLWithAnchor() = %q, want %q", got, want)
	}
}

func TestURLWithoutAnchor(t *testing.T) {
	got := URLWithoutAnchor("https://example.com/page#section")

	want := "https://example.com/page"
	if got != want {
		t.Fatalf("URLWithoutAnchor() = %q, want %q", got, want)
	}
}

func TestBreadcrumbFromURLStripsAlgoliaDocPrefix(t *testing.T) {
	got := BreadcrumbFromURL("https://www.algolia.com/doc/guides/building-search/intro#section")

	want := "/guides/building-search/intro"
	if got != want {
		t.Fatalf("BreadcrumbFromURL() = %q, want %q", got, want)
	}
}

func TestBreadcrumbFromURLStripsAlgoliaDocPrefixWithoutWWW(t *testing.T) {
	got := BreadcrumbFromURL("https://algolia.com/doc/rest-api/search/search-single-index")

	want := "/rest-api/search/search-single-index"
	if got != want {
		t.Fatalf("BreadcrumbFromURL() = %q, want %q", got, want)
	}
}

func TestBreadcrumbFromURLKeepsNonDocPath(t *testing.T) {
	got := BreadcrumbFromURL("https://example.com/page")

	want := "/page"
	if got != want {
		t.Fatalf("BreadcrumbFromURL() = %q, want %q", got, want)
	}
}
