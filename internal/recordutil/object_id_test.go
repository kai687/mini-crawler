package recordutil

import "testing"

func TestObjectIDFromURLStripsBaseURL(t *testing.T) {
	got := ObjectIDFromURL("https://www.algolia.com/doc/guides/getting-started/")

	want := "doc-guides-getting-started"
	if got != want {
		t.Fatalf("ObjectIDFromURL() = %q, want %q", got, want)
	}
}

func TestObjectIDFromURLStripsBaseURLWithoutWWW(t *testing.T) {
	got := ObjectIDFromURL("https://algolia.com/doc/guides/getting-started/")

	want := "doc-guides-getting-started"
	if got != want {
		t.Fatalf("ObjectIDFromURL() = %q, want %q", got, want)
	}
}

func TestObjectIDFromURLStripsAnyAbsoluteHost(t *testing.T) {
	got := ObjectIDFromURL("https://example.com/page")

	want := "page"
	if got != want {
		t.Fatalf("ObjectIDFromURL() = %q, want %q", got, want)
	}
}

func TestObjectIDFromURLKeepsFragment(t *testing.T) {
	got := ObjectIDFromURL("https://example.com/page#section")

	want := "page#section"
	if got != want {
		t.Fatalf("ObjectIDFromURL() = %q, want %q", got, want)
	}
}
