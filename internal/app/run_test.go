package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algolia/docs-crawler/internal/config"
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

	records := runToRecords(t, config.Config{
		Mode:   config.ModeSingle,
		Target: server.URL,
		Script: algoliaScriptPath(),
	})

	assertRecordCount(t, records, 4)
	assertRecordValue(t, records[0], "recordType", "lvl1")
	assertRecordValue(t, records[0], "content", "Hello docs")
	assertRecordValue(t, records[1], "recordType", "lvl2")
	assertRecordValue(t, records[2], "recordType", "content")
	assertRecordValue(t, records[2], "content", "First paragraph")
	assertRecordValue(t, records[3], "recordType", "lvl3")
}

func TestAlgoliaExampleObjectIDsAreUnique(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head>` +
			`<meta name="description" content="Duplicate test">` +
			`</head><body>` +
			`<h1 id="page-title">Duplicate test</h1>` +
			`<div id="content">` +
			`<h2 id="symptoms">Symptoms</h2>` +
			`<span data-as="p">First symptom</span>` +
			`<h3 id="symptoms-2">Second symptoms heading</h3>` +
			`<div class="param-head" id="param-index-name">` +
			`<div data-component-part="field-name">indexName</div>` +
			`</div>` +
			`<div class="param-head" id="param-index-name">` +
			`<div data-component-part="field-name">index_name</div>` +
			`</div>` +
			`</div>` +
			`</body></html>`))
	}))
	defer server.Close()

	records := runToRecords(t, config.Config{
		Mode:   config.ModeSingle,
		Target: server.URL + "/doc/guides/example",
		Script: algoliaScriptPath(),
	})

	seen := map[string]bool{}

	for _, record := range records {
		objectID, ok := record["objectID"].(string)
		if !ok || objectID == "" {
			t.Fatalf(
				"record objectID = %#v, want non-empty string in %#v",
				record["objectID"],
				record,
			)
		}

		if seen[objectID] {
			t.Fatalf("duplicate objectID %q in records %#v", objectID, records)
		}

		seen[objectID] = true
	}
}

func TestRunAlgoliaExampleOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(algoliaFixtureHTML()))
	}))
	defer server.Close()

	url := server.URL + "/doc/rest-api/search/search-single-index"

	records := runToRecords(t, config.Config{
		Mode:   config.ModeSingle,
		Target: url,
		Script: algoliaScriptPath(),
	})

	assertRecordCount(t, records, 7)
	assertRecordValue(t, records[0], "contentType", "api")
	assertRecordValue(t, records[0], "methodName", "searchSingleIndex")
	assertRecordValue(t, records[3], "content", "Read Reference first")
	assertRecordValue(t, records[5], "recordType", "field")
	assertRecordValue(t, records[5], "content", "string. required. Search query.")
	assertRecordValue(t, records[6], "recordType", "content")
	assertRecordValue(t, records[6], "content", "Nested after field")
}

func TestRunSingleModeWithScript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><h1 id="page-title">Script title</h1></body></html>`))
	}))
	defer server.Close()

	scriptPath := writeScript(t, `
def page_meta(doc, ctx):
    return {"title": text(doc.select_one("h1#page-title"))}

def records(doc, ctx):
    return [{"url": ctx["url"], "title": ctx["metadata"]["pageMeta"]["title"]}]

def enrich(record, ctx):
    record["position"] = ctx["position"]
    return record
`)

	var out bytes.Buffer

	err := Run(context.Background(), config.Config{
		Mode:   config.ModeSingle,
		Target: server.URL,
		Script: scriptPath,
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	want := `{"position":0,"title":"Script title","url":"` + server.URL + `"}` + "\n"
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
		Script:  algoliaScriptPath(),
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	if !strings.Contains(out.String(), pageServer.URL) {
		t.Fatalf("output = %q, want good page record", out.String())
	}
}

func runToRecords(t *testing.T, cfg config.Config) []map[string]any {
	t.Helper()

	var out bytes.Buffer
	if err := Run(context.Background(), cfg, &out); err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	var records []map[string]any

	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}

		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("Unmarshal(%q) err = %v", line, err)
		}

		records = append(records, record)
	}

	return records
}

func assertRecordCount(t *testing.T, records []map[string]any, want int) {
	t.Helper()

	if len(records) != want {
		t.Fatalf("len(records) = %d, want %d: %#v", len(records), want, records)
	}
}

func assertRecordValue(t *testing.T, record map[string]any, key string, want any) {
	t.Helper()

	if record[key] != want {
		t.Fatalf("record[%q] = %#v, want %#v in %#v", key, record[key], want, record)
	}
}

func algoliaScriptPath() string {
	return filepath.Join("..", "..", "examples", "algolia.star")
}

func algoliaFixtureHTML() string {
	return `<html><head>` +
		`<meta name="description" content="  Search one index  ">` +
		`</head><body>` +
		`<h1 id="page-title">Search single index</h1>` +
		`<div id="content">` +
		`<h2 id="overview">Overview</h2>` +
		`<span data-as="p"> First paragraph </span>` +
		`<ul>` +
		`<li><a href="/docs/ref">Reference</a></li>` +
		`<li>Read <a href="/docs/ref">Reference</a> first</li>` +
		`</ul>` +
		`<h3 id="parameters">Parameters</h3>` +
		`<div class="param-head" id="body-query">` +
		`<div data-component-part="field-name">query</div>` +
		`<div data-component-part="field-info-pill">string</div>` +
		`<div data-component-part="field-required-pill">required</div>` +
		`</div>` +
		`<div class="mt-4"><div><p>Search query.</p></div></div>` +
		`<span data-as="p"> Nested after field </span>` +
		`</div>` +
		`</body></html>`
}

func writeScript(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "script.star")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() err = %v", err)
	}

	return path
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
		Script:      algoliaScriptPath(),
	}, &out)
	if err == nil {
		t.Fatal("Run() err = nil, want error")
	}
}
