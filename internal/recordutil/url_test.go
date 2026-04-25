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

func TestBreadcrumbPathFromURLStripsAlgoliaDocPrefix(t *testing.T) {
	got := BreadcrumbPathFromURL("https://www.algolia.com/doc/guides/building-search/intro#section")

	want := "/guides/building-search/intro"
	if got != want {
		t.Fatalf("BreadcrumbPathFromURL() = %q, want %q", got, want)
	}
}

func TestBreadcrumbPathFromURLStripsAlgoliaDocPrefixWithoutWWW(t *testing.T) {
	got := BreadcrumbPathFromURL("https://algolia.com/doc/rest-api/search/search-single-index")

	want := "/rest-api/search/search-single-index"
	if got != want {
		t.Fatalf("BreadcrumbPathFromURL() = %q, want %q", got, want)
	}
}

func TestBreadcrumbSegmentsFromURL(t *testing.T) {
	got := BreadcrumbSegmentsFromURL(
		"https://www.algolia.com/doc/rest-api/search/search-single-index",
	)
	want := []string{"REST API", "Search"}

	if len(got) != len(want) {
		t.Fatalf("len(BreadcrumbSegmentsFromURL()) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("BreadcrumbSegmentsFromURL()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestBreadcrumbSegmentsFromURLUsesSentenceCaseAndPreservesSpecialCasing(t *testing.T) {
	got := BreadcrumbSegmentsFromURL(
		"https://www.algolia.com/doc/ui-libraries/what-is-autocomplete",
	)
	want := []string{"UI libraries"}

	if len(got) != len(want) {
		t.Fatalf("len(BreadcrumbSegmentsFromURL()) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("BreadcrumbSegmentsFromURL()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestBreadcrumbSegmentsFromURLReturnsNilForSingleSegmentPath(t *testing.T) {
	got := BreadcrumbSegmentsFromURL("https://www.algolia.com/doc/guides")
	if got != nil {
		t.Fatalf("BreadcrumbSegmentsFromURL() = %v, want nil", got)
	}
}

func TestMethodNameFromURL(t *testing.T) {
	got := MethodNameFromURL("https://www.algolia.com/doc/rest-api/search/search-single-index")

	want := "searchSingleIndex"
	if got != want {
		t.Fatalf("MethodNameFromURL() = %q, want %q", got, want)
	}
}

func TestMethodNameFromURLUsesSpecialCasing(t *testing.T) {
	got := MethodNameFromURL("https://www.algolia.com/doc/rest-api/search/get-ab-test")

	want := "getABTest"
	if got != want {
		t.Fatalf("MethodNameFromURL() = %q, want %q", got, want)
	}
}

func TestMethodNameFromURLSupportsSDKMethodURLs(t *testing.T) {
	got := MethodNameFromURL(
		"https://www.algolia.com/doc/libraries/sdk/methods/search/list-api-keys",
	)

	want := "listApiKeys"
	if got != want {
		t.Fatalf("MethodNameFromURL() = %q, want %q", got, want)
	}
}

func TestMethodNameFromURLReturnsEmptyForUnsupportedURL(t *testing.T) {
	got := MethodNameFromURL("https://www.algolia.com/doc/guides/building-search/intro")
	if got != "" {
		t.Fatalf("MethodNameFromURL() = %q, want empty", got)
	}
}

func TestBreadcrumbHierarchyFromSegments(t *testing.T) {
	got := BreadcrumbHierarchyFromSegments([]string{"Guides", "Building search", "Intro"})
	if got == nil {
		t.Fatal("BreadcrumbHierarchyFromSegments() = nil")
	}

	assertStringPtr(t, "Lvl0", got.Lvl0, testStringPtr("Guides"))
	assertStringPtr(t, "Lvl1", got.Lvl1, testStringPtr("Guides > Building search"))
	assertStringPtr(t, "Lvl2", got.Lvl2, testStringPtr("Guides > Building search > Intro"))
}

func testStringPtr(value string) *string { return &value }

func assertStringPtr(t *testing.T, name string, got *string, want *string) {
	t.Helper()

	if got == nil || want == nil || *got != *want {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}
