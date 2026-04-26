package output

import (
	"bytes"
	"testing"

	"github.com/algolia/docs-crawler/internal/model"
)

func TestJSONLWriterWrite(t *testing.T) {
	var buf bytes.Buffer

	writer := NewJSONLWriter(&buf)

	err := writer.Write(model.Record{
		URL:                "https://example.com/page",
		URLWithoutAnchor:   "https://example.com/page",
		BreadcrumbSegments: []string{"Guides", "Getting Started"},
		BreadcrumbHierarchy: &model.BreadcrumbHierarchy{
			Lvl0: stringPtr("Guides"),
			Lvl1: stringPtr("Guides > Getting Started"),
		},
		ContentType: "guide",
		Product:     "autocomplete",
		MethodName:  "searchSingleIndex",
		RecordType:  "content",
		Content:     stringPtr("Paragraph content"),
		Hierarchy: model.Hierarchy{
			Lvl1: stringPtr("Page Title"),
			Lvl2: stringPtr("Section"),
			Lvl3: stringPtr("Detail"),
		},
		Position: 0,
		ObjectID: "https:--example.com-page-0",
	})
	if err != nil {
		t.Fatalf("Write() err = %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() err = %v", err)
	}

	want := "{" +
		"\"url\":\"https://example.com/page\"," +
		"\"urlWithoutAnchor\":\"https://example.com/page\"," +
		"\"breadcrumbSegments\":[\"Guides\",\"Getting Started\"]," +
		"\"breadcrumbHierarchy\":{\"lvl0\":\"Guides\",\"lvl1\":\"Guides \\u003e Getting Started\"}," +
		"\"contentType\":\"guide\"," +
		"\"product\":\"autocomplete\"," +
		"\"methodName\":\"searchSingleIndex\"," +
		"\"recordType\":\"content\"," +
		"\"content\":\"Paragraph content\"," +
		"\"hierarchy\":{" +
		"\"lvl1\":\"Page Title\",\"lvl2\":\"Section\"," +
		"\"lvl3\":\"Detail\"}," +
		"\"position\":0," +
		"\"objectID\":\"https:--example.com-page-0\"" +
		"}\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestJSONLWriterWriteMultiple(t *testing.T) {
	var buf bytes.Buffer

	writer := NewJSONLWriter(&buf)

	records := []model.Record{
		{
			URL:              "https://example.com/one",
			URLWithoutAnchor: "https://example.com/one",
			RecordType:       "content",
			Content:          stringPtr("Body one"),
			Hierarchy: model.Hierarchy{
				Lvl1: stringPtr("One"),
				Lvl2: stringPtr("Section One"),
				Lvl3: stringPtr("Detail One"),
			},
			Position: 0,
			ObjectID: "https:--example.com-one-0",
		},
		{
			URL:              "https://example.com/two",
			URLWithoutAnchor: "https://example.com/two",
			RecordType:       "content",
			Content:          stringPtr("Body two"),
			Hierarchy: model.Hierarchy{
				Lvl1: stringPtr("Two"),
				Lvl2: stringPtr("Section Two"),
				Lvl3: stringPtr("Detail Two"),
			},
			Position: 1,
			ObjectID: "https:--example.com-two-1",
		},
	}
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			t.Fatalf("Write() err = %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() err = %v", err)
	}

	want := "{" +
		"\"url\":\"https://example.com/one\"," +
		"\"urlWithoutAnchor\":\"https://example.com/one\"," +
		"\"recordType\":\"content\"," +
		"\"content\":\"Body one\"," +
		"\"hierarchy\":{" +
		"\"lvl1\":\"One\",\"lvl2\":\"Section One\"," +
		"\"lvl3\":\"Detail One\"}," +
		"\"position\":0," +
		"\"objectID\":\"https:--example.com-one-0\"" +
		"}\n" +
		"{" +
		"\"url\":\"https://example.com/two\"," +
		"\"urlWithoutAnchor\":\"https://example.com/two\"," +
		"\"recordType\":\"content\"," +
		"\"content\":\"Body two\"," +
		"\"hierarchy\":{" +
		"\"lvl1\":\"Two\",\"lvl2\":\"Section Two\"," +
		"\"lvl3\":\"Detail Two\"}," +
		"\"position\":1," +
		"\"objectID\":\"https:--example.com-two-1\"" +
		"}\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func stringPtr(value string) *string {
	return &value
}
