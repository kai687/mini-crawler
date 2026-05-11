package crawl

import (
	"context"

	"github.com/kai687/mini-crawler/pkg/crawler"
	"github.com/kai687/mini-crawler/pkg/source"
	"github.com/spf13/cobra"
)

func newSingleCommand(ctx context.Context, cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "single [flags] <url>",
		Short: "Crawl one explicit page URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runCrawl(ctx, *cfg, crawler.Pipeline{
				Source:      source.Single{URL: args[0]},
				Workers:     1,
				FailOnError: true,
			})
		},
	}

	cmd.Example = "  mini-crawler crawl single " +
		"https://algolia.com/doc/ui-libraries/autocomplete/introduction/what-is-autocomplete"

	return cmd
}
