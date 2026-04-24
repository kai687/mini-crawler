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
