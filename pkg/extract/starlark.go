// Package extract contains extractor implementations for crawler pipelines.
package extract

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kai687/mini-crawler/pkg/model"
	"github.com/kai687/mini-crawler/pkg/script"
	starlarkengine "github.com/kai687/mini-crawler/pkg/script/starlark"
)

// StarlarkExtractor extracts records from parsed HTML pages with a Starlark program.
type StarlarkExtractor struct {
	Program script.Program
}

// NewStarlarkExtractor loads one Starlark script file.
func NewStarlarkExtractor(scriptPath string, debug bool) (StarlarkExtractor, error) {
	program, err := starlarkengine.Engine{}.Load(scriptPath)
	if err != nil {
		return StarlarkExtractor{}, fmt.Errorf("load script: %w", err)
	}

	if debug {
		setDebugWriter(program, os.Stderr)
	}

	return StarlarkExtractor{Program: program}, nil
}

type debugWriterProgram interface {
	SetDebugWriter(out io.Writer)
}

func setDebugWriter(program script.Program, out io.Writer) {
	if debugProgram, ok := program.(debugWriterProgram); ok {
		debugProgram.SetDebugWriter(out)
	}
}

// Extract runs the loaded Starlark program for one parsed page.
func (e StarlarkExtractor) Extract(_ context.Context, page model.ParsedPage) ([]any, error) {
	if e.Program == nil {
		return nil, fmt.Errorf("starlark program required")
	}

	doc := script.NewDocument(page)
	scriptCtx := script.Context{URL: page.URL, Metadata: page.Metadata}

	records, err := e.Program.Extract(doc, scriptCtx)
	if err != nil {
		return nil, err
	}

	out := make([]any, 0, len(records))
	for _, record := range records {
		out = append(out, record)
	}

	return out, nil
}
