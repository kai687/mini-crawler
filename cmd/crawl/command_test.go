package crawl

import (
	"strings"
	"testing"
)

func TestValidateConfigRequiresOutputForDebugScript(t *testing.T) {
	err := validateConfig(config{Script: "extract.star", DebugScript: true})
	if err == nil {
		t.Fatal("validateConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "debug-script requires --output") {
		t.Fatalf("validateConfig() error = %q, want debug-script output requirement", err)
	}
}

func TestValidateConfigAllowsDebugScriptWithOutput(t *testing.T) {
	err := validateConfig(
		config{Script: "extract.star", DebugScript: true, Output: "records.jsonl"},
	)
	if err != nil {
		t.Fatalf("validateConfig() error = %v, want nil", err)
	}
}
