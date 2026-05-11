package crawler

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

	"github.com/kai687/mini-crawler/pkg/extract"
	"github.com/kai687/mini-crawler/pkg/fetch"
	"github.com/kai687/mini-crawler/pkg/model"
	"github.com/kai687/mini-crawler/pkg/output"
	"github.com/kai687/mini-crawler/pkg/parse"
	"github.com/kai687/mini-crawler/pkg/source"
)

type testOptions struct {
	Source        Source
	Workers       int
	FailOnError   bool
	IgnoreNoindex bool
	Script        string
}

func runPipeline(ctx context.Context, opts testOptions, out *bytes.Buffer) error {
	extractor, err := extract.NewStarlarkExtractor(opts.Script, false)
	if err != nil {
		return err
	}

	var filter ParsedPageFilter
	if !opts.IgnoreNoindex {
		filter = RobotsNoindexFilter{}
	}

	return Run(ctx, Pipeline{
		Source:           opts.Source,
		Fetcher:          fetch.HTTPFetcher{},
		Parser:           parse.HTMLParser{},
		ParsedPageFilter: filter,
		Extractor:        extractor,
		Writer:           output.NewJSONLWriter(out),
		Workers:          opts.Workers,
		FailOnError:      opts.FailOnError,
	})
}

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

	records := runToRecords(t, testOptions{
		Source: source.Single{URL: server.URL + "/doc/guides/example"},
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

	records := runToRecords(t, testOptions{
		Source: source.Single{URL: server.URL + "/doc/guides/example"},
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

	records := runToRecords(t, testOptions{
		Source: source.Single{URL: url},
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
def extract_page(pattern, doc, ctx):
    return [{"url": ctx["url"], "title": text(doc.select_first("h1#page-title")), "pattern": pattern}]

extract(".*", extract_page)
`)

	var out bytes.Buffer

	err := runPipeline(context.Background(), testOptions{
		Source: source.Single{URL: server.URL},
		Script: scriptPath,
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	want := `{"pattern":".*","title":"Script title","url":"` + server.URL + `"}` + "\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestRunSkipsRobotsNoindexByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(
				`<html><head><meta name="robots" content="noindex, follow"></head><body><h1>Hidden</h1></body></html>`,
			),
		)
	}))
	defer server.Close()

	scriptPath := writeScript(t, `
def extract_page(pattern, doc, ctx):
    return [{"url": ctx["url"]}]

extract(".*", extract_page)
`)

	var out bytes.Buffer

	err := runPipeline(context.Background(), testOptions{
		Source: source.Single{URL: server.URL},
		Script: scriptPath,
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	if out.String() != "" {
		t.Fatalf("output = %q, want no records", out.String())
	}
}

func TestRunCanIgnoreRobotsNoindex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(
				`<html><head><meta name="robots" content="NOINDEX"></head><body><h1>Hidden</h1></body></html>`,
			),
		)
	}))
	defer server.Close()

	scriptPath := writeScript(t, `
def extract_page(pattern, doc, ctx):
    return [{"url": ctx["url"]}]

extract(".*", extract_page)
`)

	var out bytes.Buffer

	err := runPipeline(context.Background(), testOptions{
		Source:        source.Single{URL: server.URL},
		Script:        scriptPath,
		IgnoreNoindex: true,
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	want := `{"url":"` + server.URL + `"}` + "\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestRunSupportsCustomPipelineStages(t *testing.T) {
	var out bytes.Buffer

	err := Run(context.Background(), Pipeline{
		Source:    fakeSource{refs: []string{"docs/one.md"}},
		Fetcher:   fakeFetcher{},
		Parser:    fakeParser{},
		Extractor: fakeExtractor{},
		Writer:    output.NewJSONLWriter(&out),
	})
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	want := `{"kind":"markdown","ref":"docs/one.md","title":"Hello"}` + "\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

type fakeSource struct {
	refs []string
}

func (s fakeSource) URLs(context.Context) ([]string, error) {
	return s.refs, nil
}

type fakeFetcher struct{}

func (fakeFetcher) Fetch(_ context.Context, ref string) (model.Page, error) {
	return model.Page{Ref: ref, Body: []byte("# Hello")}, nil
}

type fakeParser struct{}

func (fakeParser) Parse(page model.Page) (model.ParsedPage, error) {
	return model.ParsedPage{
		Ref:      page.Ref,
		Kind:     "markdown",
		Document: "Hello",
	}, nil
}

type fakeExtractor struct{}

func (fakeExtractor) Extract(_ context.Context, page model.ParsedPage) ([]any, error) {
	return []any{map[string]any{
		"ref":   page.Ref,
		"kind":  page.Kind,
		"title": page.Document,
	}}, nil
}

type recordingFetcher struct {
	fetched map[string]bool
}

func (f *recordingFetcher) Fetch(_ context.Context, ref string) (model.Page, error) {
	if f.fetched == nil {
		f.fetched = map[string]bool{}
	}

	f.fetched[ref] = true

	return model.Page{Ref: ref, Body: []byte("# Hello")}, nil
}

type recordingParser struct {
	parsed map[string]bool
}

func (p *recordingParser) Parse(page model.Page) (model.ParsedPage, error) {
	if p.parsed == nil {
		p.parsed = map[string]bool{}
	}

	p.parsed[page.Ref] = true

	return fakeParser{}.Parse(page)
}

type skipRefFilter struct {
	skip string
}

func (f skipRefFilter) FilterRef(_ context.Context, ref string) error {
	if ref == f.skip {
		return ErrFiltered
	}

	return nil
}

type skipPageFilter struct {
	skip string
}

func (f skipPageFilter) FilterPage(_ context.Context, page model.Page) error {
	if page.Ref == f.skip {
		return ErrFiltered
	}

	return nil
}

func TestRunRefFilterSkipsBeforeFetch(t *testing.T) {
	fetcher := &recordingFetcher{}

	var out bytes.Buffer

	err := Run(context.Background(), Pipeline{
		Source:    fakeSource{refs: []string{"docs/keep.md", "docs/skip.md"}},
		RefFilter: skipRefFilter{skip: "docs/skip.md"},
		Fetcher:   fetcher,
		Parser:    fakeParser{},
		Extractor: fakeExtractor{},
		Writer:    output.NewJSONLWriter(&out),
	})
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	if strings.Contains(out.String(), "docs/skip.md") {
		t.Fatalf("output = %q, want skipped ref omitted", out.String())
	}

	if fetcher.fetched["docs/skip.md"] {
		t.Fatal("skipped ref was fetched")
	}

	if !fetcher.fetched["docs/keep.md"] {
		t.Fatal("kept ref was not fetched")
	}
}

func TestRunPreParseFilterSkipsBeforeParse(t *testing.T) {
	parser := &recordingParser{}

	var out bytes.Buffer

	err := Run(context.Background(), Pipeline{
		Source:         fakeSource{refs: []string{"docs/keep.md", "docs/skip.md"}},
		Fetcher:        fakeFetcher{},
		PreParseFilter: skipPageFilter{skip: "docs/skip.md"},
		Parser:         parser,
		Extractor:      fakeExtractor{},
		Writer:         output.NewJSONLWriter(&out),
	})
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	if strings.Contains(out.String(), "docs/skip.md") {
		t.Fatalf("output = %q, want skipped page omitted", out.String())
	}

	if parser.parsed["docs/skip.md"] {
		t.Fatal("skipped page was parsed")
	}

	if !parser.parsed["docs/keep.md"] {
		t.Fatal("kept page was not parsed")
	}
}

func TestNormalizedWorkers(t *testing.T) {
	got := normalizedWorkers(8)
	if got != 8 {
		t.Fatalf("normalizedWorkers() = %d, want 8", got)
	}
}

func TestNormalizedWorkersMinimum(t *testing.T) {
	got := normalizedWorkers(0)
	if got != 1 {
		t.Fatalf("normalizedWorkers() = %d, want 1", got)
	}
}

func TestRunSkipsPagesWithoutExtractor(t *testing.T) {
	pageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><h1>Page</h1></body></html>`))
	}))
	defer pageServer.Close()

	sitemapServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>` + pageServer.URL + `/doc/keep</loc></url>
  <url><loc>` + pageServer.URL + `/other/skip</loc></url>
</urlset>`))
		}),
	)
	defer sitemapServer.Close()

	scriptPath := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    return [{"url": ctx["url"]}]

extract("^/doc/", extract_docs)
`)

	var out bytes.Buffer

	err := runPipeline(context.Background(), testOptions{
		Source:      source.Sitemap{SitemapURL: sitemapServer.URL},
		FailOnError: true,
		Script:      scriptPath,
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	if !strings.Contains(out.String(), pageServer.URL+"/doc/keep") {
		t.Fatalf("output = %q, want matching page record", out.String())
	}

	if strings.Contains(out.String(), pageServer.URL+"/other/skip") {
		t.Fatalf("output = %q, want unmatched page skipped", out.String())
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
  <url><loc>` + pageServer.URL + `/doc/guides/example</loc></url>
  <url><loc>http://127.0.0.1:1/bad</loc></url>
</urlset>`))
		}),
	)
	defer sitemapServer.Close()

	var out bytes.Buffer

	err := runPipeline(context.Background(), testOptions{
		Source:  source.Sitemap{SitemapURL: sitemapServer.URL},
		Workers: 2,
		Script:  algoliaScriptPath(),
	}, &out)
	if err != nil {
		t.Fatalf("Run() err = %v", err)
	}

	if !strings.Contains(out.String(), pageServer.URL+"/doc/guides/example") {
		t.Fatalf("output = %q, want good page record", out.String())
	}
}

func runToRecords(t *testing.T, opts testOptions) []map[string]any {
	t.Helper()

	var out bytes.Buffer
	if err := runPipeline(context.Background(), opts, &out); err != nil {
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

	err := runPipeline(context.Background(), testOptions{
		Source:      source.Sitemap{SitemapURL: sitemapServer.URL},
		Workers:     2,
		FailOnError: true,
		Script:      algoliaScriptPath(),
	}, &out)
	if err == nil {
		t.Fatal("Run() err = nil, want error")
	}
}
