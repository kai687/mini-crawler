package starlark

import (
	"fmt"
	"log/slog"
	"regexp"
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
}

// Extract calls the first registered extractor whose pattern matches ctx.URL path.
func (p *Program) Extract(doc script.Document, ctx script.Context) ([]map[string]any, error) {
	path := pathFromURL(ctx.URL)

	for _, extractor := range p.extractors {
		if !extractor.regex.MatchString(path) {
			continue
		}

		slog.Debug(
			"script extractor matched",
			"script",
			p.scriptPath,
			"extractor",
			extractor.fn.Name(),
			"pattern",
			extractor.pattern,
			"url",
			ctx.URL,
		)

		start := time.Now()

		records, err := p.extractWith(extractor, doc, ctx)
		if err != nil {
			return nil, err
		}

		slog.Debug(
			"script extractor finished",
			"script",
			p.scriptPath,
			"extractor",
			extractor.fn.Name(),
			"pattern",
			extractor.pattern,
			"url",
			ctx.URL,
			"records",
			len(records),
			"duration",
			time.Since(start),
		)

		return records, nil
	}

	return nil, fmt.Errorf("script %s: %w %s", p.scriptPath, script.ErrNoExtractor, ctx.URL)
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
