package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/kai687/mini-crawler/cmd/crawl"
	scriptcmd "github.com/kai687/mini-crawler/cmd/script"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context) error {
	return newRootCommand(ctx).Execute()
}

func newRootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:           commandName(),
		Short:         "Crawl HTML pages and emit JSONL records",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(crawl.NewCommand(ctx))
	cmd.AddCommand(scriptcmd.NewCommand(ctx))

	return cmd
}

func commandName() string {
	name := filepath.Base(os.Args[0])
	if name == "." || name == string(os.PathSeparator) || name == "" {
		return "mini-crawler"
	}

	return name
}
