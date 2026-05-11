package crawl

import (
	"context"
	"net/http"
	"time"

	"github.com/algolia/mini-crawler/pkg/crawler"
	"github.com/algolia/mini-crawler/pkg/source"
	"github.com/spf13/cobra"
)

type sitemapConfig struct {
	Workers     int
	FailOnError bool
}

func newSitemapCommand(ctx context.Context, cfg *config) *cobra.Command {
	sitemapCfg := sitemapConfig{Workers: 1}
	cmd := &cobra.Command{
		Use:   "sitemap [flags] <sitemap-url>",
		Short: "Crawl all URLs discovered from a sitemap",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runCrawl(ctx, *cfg, crawler.Pipeline{
				Source: source.Sitemap{
					SitemapURL: args[0],
					Client:     &http.Client{Timeout: 15 * time.Second},
				},
				Workers:     sitemapCfg.Workers,
				FailOnError: sitemapCfg.FailOnError,
			})
		},
	}

	cmd.Flags().IntVarP(&sitemapCfg.Workers, "workers", "w", 1, "number of concurrent page workers")
	cmd.Flags().
		BoolVar(&sitemapCfg.FailOnError, "fail-on-error", false, "fail whole run when one URL cannot be crawled")

	cmd.Example = `  docs-crawler crawl sitemap https://algolia.com/doc/sitemap.xml
  docs-crawler crawl sitemap -w 8 -o records.jsonl https://algolia.com/doc/sitemap.xml`

	return cmd
}
