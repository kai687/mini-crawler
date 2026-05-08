package cmd

import (
	"context"

	"github.com/algolia/docs-crawler/internal/config"
	"github.com/spf13/cobra"
)

func newCrawlSingleCommand(ctx context.Context, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "single [flags] <url>",
		Short: "Crawl one explicit page URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg.Mode = config.ModeSingle
			cfg.Target = args[0]

			return runCrawl(ctx, *cfg)
		},
	}

	cmd.Example = "  docs-crawler crawl single " +
		"https://algolia.com/doc/ui-libraries/autocomplete/introduction/what-is-autocomplete"

	return cmd
}
