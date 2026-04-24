package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/algolia/docs-crawler/internal/config"
	"github.com/algolia/docs-crawler/internal/recordutil"
)

func TestRunSingleMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(
				`<html><head><title>Fallback</title>` +
					`<meta name="description" content="  Hello
 docs  "></head><body>` +
					`<h1 id="page-title">Hello Docs</h1>` +
					`<div id="content">` +
					`<h2 id="first-section">First section</h2>` +
					`<span data-as="p">First paragraph</span>` +
					`<h3 id="first-detail">First detail</h3>` +
					`</div>` +
					`</body></html>`,
			),
		)
	}))
	defer server.Close()

	var out bytes.Buffer

	err := Run(context.Background(), config.Config{
		Mode:   config.ModeSingle,
		Target: server.URL,
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	sectionURL := server.URL + "#first-section"

	want := "{" +
		"\"url\":\"" + server.URL + "\"," +
		"\"type\":\"lvl1\"," +
		"\"title\":\"Hello Docs\"," +
		"\"description\":\"Hello docs\"," +
		"\"hierarchy\":{" +
		"\"lvl1\":\"Hello Docs\"}," +
		"\"position\":0," +
		"\"objectID\":\"" + recordutil.ObjectIDFromURL(server.URL) + "\"" +
		"}\n" +
		"{" +
		"\"url\":\"" + sectionURL + "\"," +
		"\"type\":\"lvl2\"," +
		"\"title\":\"Hello Docs\"," +
		"\"description\":\"Hello docs\"," +
		"\"hierarchy\":{" +
		"\"lvl1\":\"Hello Docs\",\"lvl2\":\"First section\"}," +
		"\"position\":1," +
		"\"objectID\":\"" + recordutil.ObjectIDFromURL(sectionURL) + "\"" +
		"}\n" +
		"{" +
		"\"url\":\"" + sectionURL + "\"," +
		"\"type\":\"content\"," +
		"\"title\":\"Hello Docs\"," +
		"\"description\":\"Hello docs\"," +
		"\"content\":\"First paragraph\"," +
		"\"hierarchy\":{" +
		"\"lvl1\":\"Hello Docs\",\"lvl2\":\"First section\"}," +
		"\"position\":2," +
		"\"objectID\":\"" + recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(sectionURL), 2) + "\"" +
		"}\n" +
		"{" +
		"\"url\":\"" + server.URL + "#first-detail\"," +
		"\"type\":\"lvl3\"," +
		"\"title\":\"Hello Docs\"," +
		"\"description\":\"Hello docs\"," +
		"\"hierarchy\":{" +
		"\"lvl1\":\"Hello Docs\",\"lvl2\":\"First section\"," +
		"\"lvl3\":\"First detail\"}," +
		"\"position\":3," +
		"\"objectID\":\"" + recordutil.ObjectIDFromURL(server.URL+"#first-detail") + "\"" +
		"}\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestNormalizedWorkersSingleMode(t *testing.T) {
	got := normalizedWorkers(config.Config{Mode: config.ModeSingle, Workers: 8})
	if got != 1 {
		t.Fatalf("normalizedWorkers() = %d, want 1", got)
	}
}

func TestNormalizedWorkersSitemapMode(t *testing.T) {
	got := normalizedWorkers(config.Config{Mode: config.ModeSitemap, Workers: 8})
	if got != 8 {
		t.Fatalf("normalizedWorkers() = %d, want 8", got)
	}
}

func TestRunKeepsGoingByDefault(t *testing.T) {
	pageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><h1 id="page-title">Doc</h1></body></html>`))
	}))
	defer pageServer.Close()

	sitemapServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>` + pageServer.URL + `</loc></url>
  <url><loc>http://127.0.0.1:1/bad</loc></url>
</urlset>`))
		}),
	)
	defer sitemapServer.Close()

	var out bytes.Buffer

	err := Run(context.Background(), config.Config{
		Mode:    config.ModeSitemap,
		Target:  sitemapServer.URL,
		Workers: 2,
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	if !strings.Contains(out.String(), pageServer.URL) {
		t.Fatalf("output = %q, want good page record", out.String())
	}
}

func TestRunFailsOnErrorWhenConfigured(t *testing.T) {
	sitemapServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>http://127.0.0.1:1/bad</loc></url>
</urlset>`))
		}),
	)
	defer sitemapServer.Close()

	var out bytes.Buffer

	err := Run(context.Background(), config.Config{
		Mode:        config.ModeSitemap,
		Target:      sitemapServer.URL,
		Workers:     2,
		FailOnError: true,
	}, &out)
	if err == nil {
		t.Fatal("Run() err = nil, want error")
	}
}
