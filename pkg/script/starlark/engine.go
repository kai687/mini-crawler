package starlark

import (
	"fmt"

	"github.com/algolia/mini-crawler/pkg/script"
	starlarkgo "go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// defaultMaxExecutionSteps prevents accidental infinite or very expensive scripts.
const defaultMaxExecutionSteps uint64 = 10_000_000

// Engine loads Starlark extraction programs.
type Engine struct {
	// MaxExecutionSteps limits Starlark computation steps per load/call.
	// Zero uses a safe default.
	MaxExecutionSteps uint64
}

// Load reads, parses, and evaluates one Starlark script.
func (e Engine) Load(path string) (script.Program, error) {
	thread := e.newThread("load " + path)
	extractors := []extractorRegistration{}

	_, err := starlarkgo.ExecFileOptions(
		&syntax.FileOptions{},
		thread,
		path,
		nil,
		predeclaredWithExtract(&extractors),
	)
	if err != nil {
		return nil, fmt.Errorf("load starlark %s: %w", path, err)
	}

	program := &Program{
		scriptPath:        path,
		extractors:        extractors,
		maxExecutionSteps: e.maxExecutionSteps(),
	}
	if err := program.validateExports(); err != nil {
		return nil, err
	}

	return program, nil
}

// newThread creates a Starlark thread with execution limits applied.
func (e Engine) newThread(name string) *starlarkgo.Thread {
	thread := &starlarkgo.Thread{Name: name}
	thread.SetMaxExecutionSteps(e.maxExecutionSteps())

	return thread
}

// maxExecutionSteps returns configured limit or package default.
func (e Engine) maxExecutionSteps() uint64 {
	if e.MaxExecutionSteps == 0 {
		return defaultMaxExecutionSteps
	}

	return e.MaxExecutionSteps
}

// predeclaredWithExtract adds extract(pattern, fn) registration to script builtins.
func predeclaredWithExtract(extractors *[]extractorRegistration) starlarkgo.StringDict {
	values := predeclared()
	values["extract"] = starlarkgo.NewBuiltin("extract", func(
		_ *starlarkgo.Thread,
		_ *starlarkgo.Builtin,
		args starlarkgo.Tuple,
		kwargs []starlarkgo.Tuple,
	) (starlarkgo.Value, error) {
		var (
			pattern string
			fn      starlarkgo.Callable
		)

		if err := starlarkgo.UnpackArgs(
			"extract",
			args,
			kwargs,
			"pattern",
			&pattern,
			"fn",
			&fn,
		); err != nil {
			return nil, err
		}

		*extractors = append(*extractors, extractorRegistration{pattern: pattern, fn: fn})

		return starlarkgo.None, nil
	})

	return values
}
