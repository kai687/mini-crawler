package cmd

import (
	"context"

	"github.com/algolia/docs-crawler/cmd/crawl"
	scriptcmd "github.com/algolia/docs-crawler/cmd/script"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context) error {
	return newRootCommand(ctx).Execute()
}

func newRootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "docs-crawler",
		Short:         "Crawl HTML pages and emit JSONL records",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(crawl.NewCommand(ctx))
	cmd.AddCommand(scriptcmd.NewCommand(ctx))

	return cmd
}
