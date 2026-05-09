package script

import (
	"math"
	"strings"
	"testing"
)

func TestValidateJSONValueAcceptsJSONLikeValues(t *testing.T) {
	value := map[string]any{
		"nil":    nil,
		"bool":   true,
		"string": "value",
		"int":    1,
		"uint":   uint(2),
		"float":  1.5,
		"array": []any{
			"nested",
			map[string]any{"ok": true},
		},
	}

	if err := ValidateJSONValue(value); err != nil {
		t.Fatalf("ValidateJSONValue() error = %v", err)
	}
}

func TestValidateJSONValueRejectsUnsupportedValues(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr string
	}{
		{
			name:    "struct",
			value:   struct{ Name string }{Name: "bad"},
			wantErr: "unsupported value type",
		},
		{
			name:    "map with non-string key",
			value:   map[int]any{1: "bad"},
			wantErr: "unsupported value type map[int]interface {}",
		},
		{
			name:    "typed string slice",
			value:   []string{"bad"},
			wantErr: "unsupported value type []string",
		},
		{
			name:    "nan",
			value:   math.NaN(),
			wantErr: "non-finite number",
		},
		{
			name:    "infinity",
			value:   math.Inf(1),
			wantErr: "non-finite number",
		},
		{
			name: "nested unsupported",
			value: map[string]any{
				"items": []any{func() {}},
			},
			wantErr: "$.items[0]: unsupported value type func()",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateJSONValue(test.value)
			if err == nil {
				t.Fatal("ValidateJSONValue() error = nil")
			}

			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ValidateJSONValue() error = %q, want contains %q", err, test.wantErr)
			}
		})
	}
}

func TestValidateRecordRejectsNilRecord(t *testing.T) {
	err := ValidateRecord(nil)
	if err == nil {
		t.Fatal("ValidateRecord() error = nil")
	}

	if err.Error() != "record must not be nil" {
		t.Fatalf("ValidateRecord() error = %q", err)
	}
}

func TestValidateRecordsReportsIndex(t *testing.T) {
	records := []map[string]any{
		{"ok": true},
		{"bad": math.Inf(-1)},
	}

	err := ValidateRecords(records)
	if err == nil {
		t.Fatal("ValidateRecords() error = nil")
	}

	if !strings.Contains(err.Error(), "records[1]: $.bad: non-finite number") {
		t.Fatalf("ValidateRecords() error = %q", err)
	}
}
