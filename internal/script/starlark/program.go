package starlark

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/algolia/docs-crawler/internal/script"
	starlarkgo "go.starlark.net/starlark"
)

const extractorPrefix = "extract_"

type extractorRegistration struct {
	pattern string
	fn      starlarkgo.Callable
}

// Program is one loaded Starlark extraction program.
type Program struct {
	extractors        []extractorRegistration
	maxExecutionSteps uint64
}

// Extract calls the first registered extractor whose pattern matches ctx.URL path.
func (p *Program) Extract(doc script.Document, ctx script.Context) ([]map[string]any, error) {
	for _, extractor := range p.extractors {
		matched, err := regexp.MatchString(extractor.pattern, pathFromURL(ctx.URL))
		if err != nil {
			return nil, fmt.Errorf("extractor pattern %q: %w", extractor.pattern, err)
		}

		if !matched {
			continue
		}

		value, err := p.call(
			extractor.fn,
			starlarkgo.String(extractor.pattern),
			docValue(doc),
			contextValue(ctx),
		)
		if err != nil {
			return nil, err
		}

		records, err := toRecords(value)
		if err != nil {
			return nil, fmt.Errorf("%s return: %w", extractor.fn.Name(), err)
		}

		if err := script.ValidateRecords(records); err != nil {
			return nil, fmt.Errorf("%s return: %w", extractor.fn.Name(), err)
		}

		return records, nil
	}

	return nil, fmt.Errorf("%w %s", script.ErrNoExtractor, ctx.URL)
}

func (p *Program) validateExports() error {
	if len(p.extractors) == 0 {
		return fmt.Errorf("starlark script registered no extractors")
	}

	for _, extractor := range p.extractors {
		if !strings.HasPrefix(extractor.fn.Name(), extractorPrefix) {
			return fmt.Errorf(
				"extractor function %s must start with %s",
				extractor.fn.Name(),
				extractorPrefix,
			)
		}
	}

	return nil
}

func (p *Program) call(fn starlarkgo.Callable, args ...starlarkgo.Value) (starlarkgo.Value, error) {
	thread := &starlarkgo.Thread{Name: fn.Name()}
	thread.SetMaxExecutionSteps(p.maxExecutionSteps)

	value, err := starlarkgo.Call(thread, fn, starlarkgo.Tuple(args), nil)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", fn.Name(), err)
	}

	return value, nil
}
