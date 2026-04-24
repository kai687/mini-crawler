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
		URL:         "https://example.com/page",
		Type:        "content",
		Title:       stringPtr("Page Title"),
		Description: stringPtr("Page description"),
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
		"\"type\":\"content\"," +
		"\"title\":\"Page Title\"," +
		"\"description\":\"Page description\"," +
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
			URL:         "https://example.com/one",
			Type:        "content",
			Title:       stringPtr("One"),
			Description: stringPtr("First"),
			Content:     stringPtr("Body one"),
			Hierarchy: model.Hierarchy{
				Lvl1: stringPtr("One"),
				Lvl2: stringPtr("Section One"),
				Lvl3: stringPtr("Detail One"),
			},
			Position: 0,
			ObjectID: "https:--example.com-one-0",
		},
		{
			URL:         "https://example.com/two",
			Type:        "content",
			Title:       stringPtr("Two"),
			Description: stringPtr("Second"),
			Content:     stringPtr("Body two"),
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
		"\"type\":\"content\"," +
		"\"title\":\"One\"," +
		"\"description\":\"First\"," +
		"\"content\":\"Body one\"," +
		"\"hierarchy\":{" +
		"\"lvl1\":\"One\",\"lvl2\":\"Section One\"," +
		"\"lvl3\":\"Detail One\"}," +
		"\"position\":0," +
		"\"objectID\":\"https:--example.com-one-0\"" +
		"}\n" +
		"{" +
		"\"url\":\"https://example.com/two\"," +
		"\"type\":\"content\"," +
		"\"title\":\"Two\"," +
		"\"description\":\"Second\"," +
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
