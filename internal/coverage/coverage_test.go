package coverage

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/internal/extract"
	"github.com/algolia/docs-crawler/internal/model"
)

func TestAnalyzeFullCoverage(t *testing.T) {
	page := parsedPage(t, "https://example.com/page", htmlFixture())

	records, err := extract.PageExtractor{}.Extract(page)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	report := Analyze(page, records)
	if report.ExpectedNodes != 6 {
		t.Fatalf("ExpectedNodes = %d, want 6", report.ExpectedNodes)
	}

	if report.MatchedNodes != 6 {
		t.Fatalf("MatchedNodes = %d, want 6", report.MatchedNodes)
	}

	if report.Coverage != 1 {
		t.Fatalf("Coverage = %v, want 1", report.Coverage)
	}

	if report.HeadingCoverage != 1 {
		t.Fatalf("HeadingCoverage = %v, want 1", report.HeadingCoverage)
	}

	if report.MetadataCoverage != 1 {
		t.Fatalf("MetadataCoverage = %v, want 1", report.MetadataCoverage)
	}

	if report.HierarchyCoverage != 1 {
		t.Fatalf("HierarchyCoverage = %v, want 1", report.HierarchyCoverage)
	}

	if len(report.Missing) != 0 {
		t.Fatalf("Missing = %#v, want empty", report.Missing)
	}

	if len(report.Errors) != 0 {
		t.Fatalf("Errors = %#v, want empty", report.Errors)
	}
}

func TestAnalyzeDetectsMissingHeading(t *testing.T) {
	page := parsedPage(t, "https://example.com/page", htmlFixture())

	records, err := extract.PageExtractor{}.Extract(page)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	records = records[:len(records)-1]

	report := Analyze(page, records)
	if report.HeadingCoverage >= 1 {
		t.Fatalf("HeadingCoverage = %v, want < 1", report.HeadingCoverage)
	}

	if len(report.Missing) == 0 {
		t.Fatal("Missing empty, want missing heading")
	}
}

func TestAnalyzeDetectsHierarchyMismatch(t *testing.T) {
	page := parsedPage(t, "https://example.com/page", htmlFixture())

	records, err := extract.PageExtractor{}.Extract(page)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	wrong := "Wrong"
	records[2].Hierarchy.Lvl3 = &wrong

	report := Analyze(page, records)
	if report.HierarchyCoverage >= 1 {
		t.Fatalf("HierarchyCoverage = %v, want < 1", report.HierarchyCoverage)
	}

	if len(report.Errors) == 0 {
		t.Fatal("Errors empty, want hierarchy error")
	}
}

func parsedPage(t *testing.T, pageURL string, html string) model.ParsedPage {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() err = %v", err)
	}

	return model.ParsedPage{URL: pageURL, Doc: doc}
}

func htmlFixture() string {
	return `<html><head><title>Page Title</title><meta name="description" content="Page description"></head><body>` +
		`<h1 id="page-title">Page Title</h1>` +
		`<div id="content">` +
		`<h2 id="section-one">Section one</h2>` +
		`<h3 id="detail-one">Detail one</h3>` +
		`<h4 id="subdetail-one">Subdetail one</h4>` +
		`<h2 id="section-two">Section two</h2>` +
		`</div></body></html>`
}
