package crawl

import (
	"context"
	"time"

	"github.com/kai687/mini-crawler/pkg/crawler"
	"github.com/kai687/mini-crawler/pkg/parse"
	"github.com/kai687/mini-crawler/pkg/source"
	"github.com/spf13/cobra"
)

type llmsConfig struct {
	Workers     int
	FailOnError bool
}

func newLLMSCommand(ctx context.Context, cfg *config) *cobra.Command {
	llmsCfg := llmsConfig{Workers: 1}
	cmd := &cobra.Command{
		Use:     "llms.txt [flags] <llms-url>",
		Aliases: []string{"llms"},
		Short:   "Crawl all Markdown URLs discovered from an llms.txt file",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runCrawl(ctx, *cfg, crawler.Pipeline{
				Source: source.LLMS{
					LLMSURL: args[0],
					Client:  newHTTPClient(15 * time.Second),
				},
				Parser:      parse.MarkdownParser{},
				Workers:     llmsCfg.Workers,
				FailOnError: llmsCfg.FailOnError,
			})
		},
	}

	cmd.Flags().IntVarP(&llmsCfg.Workers, "workers", "w", 1, "number of concurrent page workers")
	cmd.Flags().BoolVar(
		&llmsCfg.FailOnError,
		"fail-on-error",
		false,
		"fail whole run when one URL cannot be crawled",
	)

	cmd.Example = `  mini-crawler crawl llms.txt -s markdown.star https://example.com/llms.txt
  mini-crawler crawl llms -w 8 -o records.jsonl -s markdown.star https://example.com/llms.txt`

	return cmd
}
