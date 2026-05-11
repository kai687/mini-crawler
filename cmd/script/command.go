package script

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	starlarkengine "github.com/kai687/mini-crawler/pkg/script/starlark"
	"github.com/spf13/cobra"
)

type config struct {
	Script string
	JSON   bool
}

func NewCommand(_ context.Context) *cobra.Command {
	cfg := config{}

	cmd := &cobra.Command{
		Use:   "script",
		Short: "Inspect extraction scripts",
	}

	check := &cobra.Command{
		Use:   "check --script <script.star>",
		Short: "Load a Starlark script and list registered extractors",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCheck(cmd.OutOrStdout(), cfg)
		},
	}
	check.Flags().StringVar(&cfg.Script, "script", "", "Starlark script to check")
	check.Flags().BoolVar(&cfg.JSON, "json", false, "write script info as JSON")

	cmd.AddCommand(check)

	return cmd
}

func runCheck(out io.Writer, cfg config) error {
	if cfg.Script == "" {
		return fmt.Errorf("invalid config: script flag required")
	}

	info, err := starlarkengine.Check(cfg.Script)
	if err != nil {
		return fmt.Errorf("check script: %w", err)
	}

	if cfg.JSON {
		return writeJSON(out, info)
	}

	fmt.Fprintf(out, "script: %s\n", info.Path)
	fmt.Fprintf(out, "extractors: %d\n", len(info.Extractors))

	for i, extractor := range info.Extractors {
		fmt.Fprintf(out, "  %d. %s %q\n", i+1, extractor.Name, extractor.Pattern)
	}

	return nil
}

func writeJSON(out io.Writer, info starlarkengine.Info) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")

	return encoder.Encode(info)
}
