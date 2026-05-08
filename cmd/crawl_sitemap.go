package cmd

import (
	"context"

	"github.com/algolia/docs-crawler/internal/config"
	"github.com/spf13/cobra"
)

func newCrawlSitemapCommand(ctx context.Context, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sitemap [flags] <sitemap-url>",
		Short: "Crawl all URLs discovered from a sitemap",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg.Mode = config.ModeSitemap
			cfg.Target = args[0]

			return runCrawl(ctx, *cfg)
		},
	}

	cmd.Flags().IntVar(&cfg.Workers, "workers", 1, "number of concurrent page workers")
	cmd.Flags().StringVar(&cfg.Filter, "filter", "", "substring filter for sitemap URLs")
	cmd.Flags().
		BoolVar(&cfg.FailOnError, "fail-on-error", false, "fail whole run when one URL cannot be crawled")

	cmd.Example = `  docs-crawler crawl sitemap https://algolia.com/doc/sitemap.xml
  docs-crawler crawl sitemap --workers 8 --output records.jsonl https://algolia.com/doc/sitemap.xml
  docs-crawler crawl sitemap --filter /rest-api/ https://algolia.com/doc/sitemap.xml`

	return cmd
}
