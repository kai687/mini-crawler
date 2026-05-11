# AGENTS.md

## Commands

- Use Go `1.26.x` (`go.mod` says `go 1.26.0`; `mise.toml` installs `go = "latest"`, so do not assume older Go compatibility).
- Single package / single test: `go test ./internal/extract -run TestName`
- Repo task runner is `mise`
- Validate changes : `mise all`

## Architecture

- CLI entrypoint: `main.go` -> `internal/app.Run`
- Crawl pipeline in `internal/app/run.go`: `source` -> `fetch` -> `parse` -> `extract` -> `enrich` -> `output`
- URL discovery lives in `internal/source`:
  - `Single` returns one explicit URL
  - `Sitemap` fetches XML sitemap and resolves `<loc>` values; Starlark extractor patterns decide what gets records
- Record shaping is split on purpose:
  - DOM extraction rules in `internal/extract/page.go`
  - Algolia record/schema enrichment in `internal/enrich/records.go`
  - URL/breadcrumb/objectID/methodName helpers in `internal/recordutil`

## Behavior That Is Easy To Miss

- This is not generic crawler. Extraction is hard-coded to Algolia docs DOM selectors in `internal/extract/page.go`:
  - content root: first of `#content`, `#content-area`
  - headings: `h2[id]`..`h6[id]`
  - prose: `span[data-as='p']`
  - list items: `li` after stripping links-only content
  - API fields: `div.param-head[id]`
- Page metadata is also selector-based: title from `h1#page-title`, description from `meta[name='description']`.
- `crawl single` ignores workers; `internal/app.normalizedWorkers` always forces single-page crawls to 1 worker.
- Sitemap runs are best-effort by default. Without `--fail-on-error`, bad pages are logged and crawl continues.
- Single URL runs fail if the page cannot be crawled.
- `--verbose` writes to stderr; JSONL records go to stdout unless `--output` is used.
- `output.JSONLWriter` buffers writes and flushes on close. Keep close/flush behavior intact if touching output paths.

## Record / Indexing Assumptions

- Output schema is tailored to Algolia indexing, not generic export.
- Page record always emitted first with `position = 0`; later records use DOM order.
- `urlWithoutAnchor` is required for distinct-per-page behavior.
- `methodName` is derived only for `/rest-api/...` and `/libraries/sdk/methods/...` URLs.
- `contentTypeFromURL` currently recognizes more prefixes than the README summary: `guide`, `api`, `integration`, `sdk`.
- Field records are intentionally forced into `hierarchy.lvl3` semantics in `internal/enrich/records.go`.

## Tests

- Current tests are pure unit/integration-with-`httptest`; no external services needed for `go test ./...`.
- Highest-signal test files when changing behavior:
  - `internal/app/run_test.go` for end-to-end JSONL output
  - `internal/extract/page_test.go` for DOM selector behavior
  - `internal/enrich/records_test.go` for record hierarchy/content rules
  - `internal/recordutil/*_test.go` for URL/objectID/methodName logic

## Repo Quirks

- Root `mini-crawler` file is local build artifact; `.gitignore` ignores it.
- `mise run indexing` and `mise run set-settings` are operational tasks, not normal verification:
  - they require local `algolia` CLI
  - `indexing` also requires interactive `gum confirm`
  - `indexing` uses `./mini-crawler crawl sitemap -workers 16 https://www.algolia.com/doc/sitemap.xml | algolia objects import docs_clean -F -`
