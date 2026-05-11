package cmd

import (
	"bytes"
	"context"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := newRootCommand(context.Background())
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() err = %v", err)
	}

	want := "mini-crawler dev\ncommit: unknown\nbuilt: unknown\n"
	if got := buf.String(); got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
}
