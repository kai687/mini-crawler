package starlark

import (
	"fmt"

	"github.com/algolia/docs-crawler/internal/script"
	starlarkgo "go.starlark.net/starlark"
)

const (
	pageMetaFunc = "page_meta"
	recordsFunc  = "records"
	enrichFunc   = "enrich"
)

// Program is one loaded Starlark extraction program.
type Program struct {
	globals           starlarkgo.StringDict
	maxExecutionSteps uint64
}

// PageMeta calls the script page_meta(doc, ctx) function.
func (p *Program) PageMeta(doc script.Document, ctx script.Context) (map[string]any, error) {
	value, err := p.call(pageMetaFunc, docValue(doc), contextValue(ctx))
	if err != nil {
		return nil, err
	}

	meta, err := toStringAnyMap(value)
	if err != nil {
		return nil, fmt.Errorf("%s return: %w", pageMetaFunc, err)
	}

	if err := script.ValidateRecord(meta); err != nil {
		return nil, fmt.Errorf("%s return: %w", pageMetaFunc, err)
	}

	return meta, nil
}

// Records calls the script records(doc, ctx) function.
func (p *Program) Records(doc script.Document, ctx script.Context) ([]map[string]any, error) {
	value, err := p.call(recordsFunc, docValue(doc), contextValue(ctx))
	if err != nil {
		return nil, err
	}

	records, err := toRecords(value)
	if err != nil {
		return nil, fmt.Errorf("%s return: %w", recordsFunc, err)
	}

	if err := script.ValidateRecords(records); err != nil {
		return nil, fmt.Errorf("%s return: %w", recordsFunc, err)
	}

	return records, nil
}

// Enrich calls the script enrich(record, ctx) function.
func (p *Program) Enrich(record map[string]any, ctx script.Context) (map[string]any, error) {
	value, err := p.call(enrichFunc, fromGoValue(record), contextValue(ctx))
	if err != nil {
		return nil, err
	}

	enriched, err := toStringAnyMap(value)
	if err != nil {
		return nil, fmt.Errorf("%s return: %w", enrichFunc, err)
	}

	if err := script.ValidateRecord(enriched); err != nil {
		return nil, fmt.Errorf("%s return: %w", enrichFunc, err)
	}

	return enriched, nil
}

func (p *Program) validateExports() error {
	for _, name := range []string{pageMetaFunc, recordsFunc, enrichFunc} {
		value, ok := p.globals[name]
		if !ok {
			return fmt.Errorf("starlark script missing %s", name)
		}

		if _, ok := value.(starlarkgo.Callable); !ok {
			return fmt.Errorf("starlark export %s is not callable", name)
		}
	}

	return nil
}

func (p *Program) call(name string, args ...starlarkgo.Value) (starlarkgo.Value, error) {
	fn := p.globals[name]
	thread := &starlarkgo.Thread{Name: name}
	thread.SetMaxExecutionSteps(p.maxExecutionSteps)

	value, err := starlarkgo.Call(thread, fn, starlarkgo.Tuple(args), nil)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", name, err)
	}

	return value, nil
}
