// Package extract contains extractor implementations for crawler pipelines.
package extract

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/algolia/docs-crawler/pkg/model"
	"github.com/algolia/docs-crawler/pkg/script"
	starlarkengine "github.com/algolia/docs-crawler/pkg/script/starlark"
)

// StarlarkExtractor extracts records from parsed HTML pages with a Starlark program.
type StarlarkExtractor struct {
	Program script.Program
	Debug   bool
}

// NewStarlarkExtractor loads one Starlark script file.
func NewStarlarkExtractor(scriptPath string, debug bool) (StarlarkExtractor, error) {
	program, err := starlarkengine.Engine{}.Load(scriptPath)
	if err != nil {
		return StarlarkExtractor{}, fmt.Errorf("load script: %w", err)
	}

	return StarlarkExtractor{Program: program, Debug: debug}, nil
}

// Extract runs the loaded Starlark program for one parsed page.
func (e StarlarkExtractor) Extract(_ context.Context, page model.ParsedPage) ([]any, error) {
	if e.Program == nil {
		return nil, fmt.Errorf("starlark program required")
	}

	doc := script.NewDocument(page)
	scriptCtx := script.Context{URL: page.URL, Metadata: page.Metadata}

	start := time.Now()

	records, err := e.Program.Extract(doc, scriptCtx)
	if err != nil {
		return nil, err
	}

	if e.Debug {
		slog.Debug(
			"script page extracted",
			"url",
			page.URL,
			"records",
			len(records),
			"duration",
			time.Since(start),
		)
	}

	out := make([]any, 0, len(records))
	for _, record := range records {
		out = append(out, record)
	}

	return out, nil
}
