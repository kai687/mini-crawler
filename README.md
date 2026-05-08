# docs-crawler

Small Go CLI for crawling documentation pages and emitting JSONL records.

Crawler core is generic. Site-specific extraction and record shaping live in a required Starlark script.

It supports two input modes:

- `sitemap`: load page URLs from a sitemap, then crawl each page
- `single`: crawl one explicit page URL

For each page, crawler:

1. Fetches HTML
1. Parses document
1. Runs Starlark extraction script
1. Writes one JSON object per returned record

## Requirements

- Go `1.26+`

## Build

```bash
mise build
```

## Usage

```bash
docs-crawler crawl sitemap --script <script.star> [flags] <sitemap-url>
docs-crawler crawl single --script <script.star> [flags] <url>
```

### Flags

Common `crawl` flags:

- `--script`: required Starlark script for site-specific extraction
- `--verbose`: show crawl logs on stderr
- `--output`: write JSONL to file instead of stdout

`sitemap` flags:

- `--workers`: number of concurrent page workers. Default: `1`
- `--filter`: substring filter applied to sitemap URLs before crawling
- `--fail-on-error`: stop run if one URL fails

## Examples

Crawl one page with Algolia example script:

```bash
go run . crawl single \
  --script examples/algolia.star \
  https://algolia.com/doc/ui-libraries/autocomplete/introduction/what-is-autocomplete
```

Crawl sitemap with 8 workers and save output:

```bash
go run . crawl sitemap \
  --script examples/algolia.star \
  --workers 8 \
  --output records.jsonl \
  https://algolia.com/sitemap.xml
```

Only crawl sitemap URLs containing `/rest-api/`:

```bash
go run . crawl sitemap \
  --script examples/algolia.star \
  --filter /rest-api/ \
  --output records.jsonl \
  https://algolia.com/sitemap.xml
```

Fail fast on first page error:

```bash
go run . crawl sitemap \
  --script examples/algolia.star \
  --workers 4 \
  --fail-on-error \
  https://algolia.com/sitemap.xml
```

## Starlark script contract

Scripts must export three functions:

```python
def page_meta(doc, ctx):
    return {}


def records(doc, ctx):
    return []


def enrich(record, ctx):
    return record
```

Return values must be JSON-like:

- `None`
- booleans
- strings
- finite numbers
- lists
- dicts with string keys

Crawler calls functions in this order per page:

1. `page_meta(doc, ctx)`
1. `records(doc, ctx)` with `ctx["metadata"]["pageMeta"]` set to page meta
1. `enrich(record, ctx)` for each record returned by `records`

`ctx["position"]` in `enrich` is the zero-based index of the record returned by `records`.

### `doc`

`doc` exposes safe DOM methods:

- `doc.url`
- `doc.select(css)` -> list of nodes
- `doc.select_one(css)` -> node or `None`
- `node.select(css)` -> list of descendant nodes
- `node.select_one(css)` -> descendant node or `None`
- `node.next(css)` -> next sibling matching CSS or `None`

### Host helpers

DOM helpers:

- `text(node)`
- `attr(node, name)` -> string or `None`
- `node_name(node)`
- `has_parent(node, css)`
- `clone_without_text(node, css)` -> clone node, remove matching descendants, return text

String helpers:

- `trim(s)`
- `collapse_space(s)`

URL helpers:

- `url_join(base, ref)`
- `url_without_anchor(url)`
- `path(url)`

Other helpers:

- `sha1(s)`
- `regex_match(pattern, s)`
- `regex_replace(pattern, repl, s)`

## Output

Output is newline-delimited JSON (`.jsonl`). Each line is one script-produced record after `enrich`.

The crawler does not enforce a record schema beyond JSON-like values. Your script owns fields such as `url`, `objectID`, `recordType`, `hierarchy`, and `content`.

Example line from `examples/algolia.star`:

```json
{
  "url": "https://algolia.com/doc/rest-api/search#query",
  "urlWithoutAnchor": "https://algolia.com/doc/rest-api/search",
  "breadcrumbSegments": ["REST API"],
  "breadcrumbHierarchy": { "lvl0": "REST API" },
  "contentType": "api",
  "recordType": "field",
  "content": "string. required. Search query text.",
  "hierarchy": { "lvl1": "Search", "lvl3": "query" },
  "position": 12,
  "objectID": "doc-rest-api-search-query"
}
```

## Algolia example script

`examples/algolia.star` ports the old hard-coded Algolia docs behavior into Starlark.

It extracts content under first matching root:

- `#content`
- `#content-area`

Inside content root, it extracts:

- headings: `h2[id]` through `h6[id]`
- paragraph-like text: `span[data-as='p']`
- list items: `li` with non-empty text after links are stripped
- API field headers: `div.param-head[id]`

Page-level metadata:

- title: `h1#page-title`
- description: `meta[name='description']`

It also builds Algolia-oriented fields:

- `url`
- `urlWithoutAnchor`
- `breadcrumbSegments`
- `breadcrumbHierarchy`
- `contentType`
- `variant`
- `methodName`
- `recordType`
- `content`
- `hierarchy`
- `position`
- `objectID`

Those fields are script behavior, not crawler core behavior.

## Error policy

Single URL runs fail if the page cannot be crawled.

Sitemap runs are best-effort by default: failed pages are logged and crawl continues. Use `--fail-on-error` to stop on first page failure.

Verbose logs go to stderr. JSONL records go to stdout unless `--output` is set.

## Project layout

- `main.go`: CLI entrypoint
- `cmd`: CLI commands and flags
- `internal/app`: crawl pipeline orchestration
- `internal/config`: CLI config validation
- `internal/source`: URL discovery for single URL or sitemap
- `internal/fetch`: HTTP page fetching
- `internal/parse`: HTML parsing
- `internal/script`: language-neutral script interfaces and JSON validation
- `internal/script/starlark`: Starlark engine and host API
- `internal/output`: JSONL writer
- `examples/algolia.star`: Algolia docs extraction/enrichment script

## Test

```bash
go test ./...
```

Full validation:

```bash
mise all
```
