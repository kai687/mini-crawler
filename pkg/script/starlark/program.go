package starlark

import (
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/kai687/mini-crawler/pkg/script"
	starlarkgo "go.starlark.net/starlark"
)

// extractorRegistration stores one extract(pattern, fn) call from a script.
type extractorRegistration struct {
	pattern string
	regex   *regexp.Regexp
	fn      starlarkgo.Callable
}

// Program is one loaded Starlark extraction program.
type Program struct {
	scriptPath        string
	extractors        []extractorRegistration
	maxExecutionSteps uint64
	debugOut          io.Writer
	debugMu           sync.Mutex
}

// SetDebugWriter enables human-readable script debug logs.
func (p *Program) SetDebugWriter(out io.Writer) {
	p.debugOut = out
}

// Extract calls the first registered extractor whose pattern matches ctx.URL path.
func (p *Program) Extract(doc script.Document, ctx script.Context) ([]map[string]any, error) {
	path := pathFromURL(ctx.URL)

	for _, extractor := range p.extractors {
		if !extractor.regex.MatchString(path) {
			continue
		}

		p.writeDebug(
			"extractor matched",
			"url", ctx.URL,
			"extractor", extractor.fn.Name(),
			"pattern", extractor.pattern,
			"script", p.scriptPath,
		)

		start := time.Now()

		records, err := p.extractWith(extractor, doc, ctx)
		if err != nil {
			return nil, err
		}

		p.writeDebug(
			"extractor finished",
			"url", ctx.URL,
			"extractor", extractor.fn.Name(),
			"pattern", extractor.pattern,
			"script", p.scriptPath,
			"records", len(records),
			"duration", time.Since(start).Round(time.Millisecond),
		)

		return records, nil
	}

	return nil, fmt.Errorf("script %s: %w %s", p.scriptPath, script.ErrNoExtractor, ctx.URL)
}

func (p *Program) writeDebug(title string, fields ...any) {
	if p.debugOut == nil {
		return
	}

	p.debugMu.Lock()
	defer p.debugMu.Unlock()

	_, _ = fmt.Fprintf(p.debugOut, "debug script: %s\n", title)
	for i := 0; i+1 < len(fields); i += 2 {
		_, _ = fmt.Fprintf(p.debugOut, "  %-9s %v\n", fmt.Sprint(fields[i])+":", fields[i+1])
	}

	_, _ = fmt.Fprintln(p.debugOut)
}

func (p *Program) extractWith(
	extractor extractorRegistration,
	doc script.Document,
	ctx script.Context,
) ([]map[string]any, error) {
	value, err := p.call(
		extractor.fn,
		starlarkgo.String(extractor.pattern),
		docValue(doc),
		contextValue(ctx),
	)
	if err != nil {
		return nil, p.extractorError(extractor, ctx, err)
	}

	records, err := toRecords(value)
	if err != nil {
		return nil, p.extractorReturnError(extractor, ctx, err)
	}

	if err := script.ValidateRecords(records); err != nil {
		return nil, p.extractorReturnError(extractor, ctx, err)
	}

	return records, nil
}

func (p *Program) extractorError(
	extractor extractorRegistration,
	ctx script.Context,
	err error,
) error {
	return fmt.Errorf(
		"script %s extractor %s pattern %q url %s: %w",
		p.scriptPath,
		extractor.fn.Name(),
		extractor.pattern,
		ctx.URL,
		err,
	)
}

func (p *Program) extractorReturnError(
	extractor extractorRegistration,
	ctx script.Context,
	err error,
) error {
	return fmt.Errorf(
		"script %s extractor %s pattern %q url %s return: %w",
		p.scriptPath,
		extractor.fn.Name(),
		extractor.pattern,
		ctx.URL,
		err,
	)
}

// validateExports rejects scripts that forgot to register extractors or used invalid extractors.
func (p *Program) validateExports() error {
	if len(p.extractors) == 0 {
		return fmt.Errorf("starlark script registered no extractors")
	}

	for i := range p.extractors {
		extractor := &p.extractors[i]

		regex, err := regexp.Compile(extractor.pattern)
		if err != nil {
			return fmt.Errorf(
				"extractor %s pattern %q: %w",
				extractor.fn.Name(),
				extractor.pattern,
				err,
			)
		}

		extractor.regex = regex

		if err := validateExtractorFunction(extractor.fn); err != nil {
			return fmt.Errorf("extractor %s: %w", extractor.fn.Name(), err)
		}
	}

	return nil
}

func validateExtractorFunction(fn starlarkgo.Callable) error {
	function, ok := fn.(*starlarkgo.Function)
	if !ok {
		return fmt.Errorf("must be a starlark function, got %s", fn.Type())
	}

	if function.HasVarargs() || function.HasKwargs() || function.NumKwonlyParams() != 0 {
		return fmt.Errorf("must accept exactly 3 positional arguments")
	}

	if function.NumParams() != 3 {
		return fmt.Errorf(
			"must accept exactly 3 positional arguments, got %d",
			function.NumParams(),
		)
	}

	return nil
}

// call invokes one extractor with execution limits applied.
func (p *Program) call(fn starlarkgo.Callable, args ...starlarkgo.Value) (starlarkgo.Value, error) {
	thread := &starlarkgo.Thread{Name: fn.Name()}
	thread.SetMaxExecutionSteps(p.maxExecutionSteps)

	value, err := starlarkgo.Call(thread, fn, starlarkgo.Tuple(args), nil)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", fn.Name(), err)
	}

	return value, nil
}
