package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// Execute runs the docs-crawler CLI.
func Execute(ctx context.Context) error {
	return newRootCommand(ctx).Execute()
}

func newRootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "docs-crawler",
		Short:         "Crawl documentation pages and emit Algolia-ready JSONL records",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newCrawlCommand(ctx))

	return cmd
}
