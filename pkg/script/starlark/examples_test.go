package starlark

import (
	"path/filepath"
	"testing"
)

func TestExamplesLoad(t *testing.T) {
	examples := []string{
		"algolia.star",
		"minimal.star",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			path := filepath.Join("..", "..", "..", "examples", example)
			if _, err := (Engine{}).Load(path); err != nil {
				t.Fatalf("Load(%q) error = %v", path, err)
			}
		})
	}
}
