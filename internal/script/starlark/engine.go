package starlark

import (
	"fmt"

	"github.com/algolia/docs-crawler/internal/script"
	starlarkgo "go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

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

	globals, err := starlarkgo.ExecFileOptions(
		&syntax.FileOptions{},
		thread,
		path,
		nil,
		predeclared(),
	)
	if err != nil {
		return nil, fmt.Errorf("load starlark %s: %w", path, err)
	}

	program := &Program{
		globals:           globals,
		maxExecutionSteps: e.maxExecutionSteps(),
	}
	if err := program.validateExports(); err != nil {
		return nil, err
	}

	return program, nil
}

func (e Engine) newThread(name string) *starlarkgo.Thread {
	thread := &starlarkgo.Thread{Name: name}
	thread.SetMaxExecutionSteps(e.maxExecutionSteps())

	return thread
}

func (e Engine) maxExecutionSteps() uint64 {
	if e.MaxExecutionSteps == 0 {
		return defaultMaxExecutionSteps
	}

	return e.MaxExecutionSteps
}
