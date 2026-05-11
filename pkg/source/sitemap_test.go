package source

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestSitemapURLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != "mini-crawler/0.1" {
			t.Fatalf("User-Agent = %q", r.UserAgent())
		}

		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>/docs/intro</loc></url>
  <url><loc>https://example.org/guide</loc></url>
</urlset>`))
	}))
	defer server.Close()

	s := Sitemap{
		SitemapURL: server.URL + "/sitemap.xml",
		Client:     server.Client(),
	}

	urls, err := s.URLs(context.Background())
	if err != nil {
		t.Fatalf("URLs() err = %v", err)
	}

	want := []string{
		server.URL + "/docs/intro",
		"https://example.org/guide",
	}
	if !reflect.DeepEqual(urls, want) {
		t.Fatalf("URLs() = %#v, want %#v", urls, want)
	}
}

func TestSitemapIndexURLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != "mini-crawler/0.1" {
			t.Fatalf("User-Agent = %q", r.UserAgent())
		}

		w.Header().Set("Content-Type", "application/xml")

		switch r.URL.Path {
		case "/sitemap.xml":
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap><loc>/docs-sitemap.xml</loc></sitemap>
  <sitemap><loc>/api-sitemap.xml</loc></sitemap>
</sitemapindex>`))
		case "/docs-sitemap.xml":
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>/docs/intro</loc></url>
</urlset>`))
		case "/api-sitemap.xml":
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>/api/search</loc></url>
</urlset>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	s := Sitemap{
		SitemapURL: server.URL + "/sitemap.xml",
		Client:     server.Client(),
	}

	urls, err := s.URLs(context.Background())
	if err != nil {
		t.Fatalf("URLs() err = %v", err)
	}

	want := []string{
		server.URL + "/docs/intro",
		server.URL + "/api/search",
	}
	if !reflect.DeepEqual(urls, want) {
		t.Fatalf("URLs() = %#v, want %#v", urls, want)
	}
}

func TestResolveURL(t *testing.T) {
	base := mustParseURL(t, "https://example.com/sitemap.xml")

	got, err := resolveURL(base, "/docs/page")
	if err != nil {
		t.Fatalf("resolveURL() err = %v", err)
	}

	if got != "https://example.com/docs/page" {
		t.Fatalf("resolveURL() = %q", got)
	}
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse() err = %v", err)
	}

	return u
}
