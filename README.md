# docs-crawler

Small Go library and CLI for turning documentation sources into JSONL records.

## Mental model

`docs-crawler` is a pipeline runner, not a scraper framework with one fixed idea of a page.
A crawl is five replaceable stages:

```text
Source -> Fetcher -> Parser -> Extractor -> Writer
```

- **Source** discovers references to process. A reference can be an HTTP URL, file path, object key, database ID, or any string your fetcher understands.
- **Fetcher** loads raw bytes for one reference.
- **Parser** turns raw bytes into a parsed document shape.
- **Extractor** reads the parsed document and returns JSON-like records.
- **Writer** receives records and persists them.

The CLI is one default assembly of this pipeline:

```text
sitemap/single URL source -> HTTP fetcher -> HTML parser -> Starlark extractor -> JSONL writer
```

Library users can swap any stage. For example, local Markdown docs could use:

```text
filesystem source -> file fetcher -> Markdown parser -> custom extractor -> JSONL writer
```

The crawler core only coordinates discovery, worker fan-out, error policy, and record writing. It doesn't know about Algolia records, HTML selectors, Markdown, or HTTP beyond the concrete stages you plug in.

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

Fail fast on first page error:

```bash
go run . crawl sitemap \
  --script examples/algolia.star \
  --workers 4 \
  --fail-on-error \
  https://algolia.com/sitemap.xml
```

## Starlark script contract

Scripts register extractor functions with the `extract(pattern, fn)` DSL:

```python
def extract_guides(pattern, doc, ctx):
    return [{"url": ctx["url"], "title": text(doc.select_first("h1"))}]


extract("^/doc/guides/", extract_guides)
```

Extractor rules:

- function may have any name
- function signature must accept exactly 3 positional arguments; names are local to the function
- `pattern` is a valid regular expression matched against the URL path
- registration order matters; first matching extractor wins
- extractor returns a list of JSON-like records
- if no extractor matches, the page is skipped with a warning

Record values must be JSON-like:

- `None`
- booleans
- strings
- finite numbers
- lists
- dicts with string keys

### `doc`

`doc` exposes safe DOM methods:

- `doc.url`
- `doc.select(css)` -> list of nodes
- `doc.select_first(css)` -> node or `None`
- `node.select(css)` -> list of descendant nodes
- `node.select_first(css)` -> descendant node or `None`
- `node.next(css)` -> next sibling matching CSS or `None`

### `ctx`

`ctx` exposes page context:

- `ctx["url"]`: current page URL
- `ctx["position"]`: zero unless your code sets it elsewhere
- `ctx["metadata"]`: optional host metadata; empty by default

### Host helpers

| Helper | Returns | Notes |
| --- | --- | --- |
| `text(node)` | string | Visible text for a node. Errors if node is `None`. |
| `safe_text(node)` | string | Visible text, or `""` when node is `None`. |
| `first_text(root, css)` | string | Text for first matching descendant of `doc` or `node`, or `""`. |
| `attr(node, name)` | string or `None` | Attribute value for a node. |
| `first_attr(root, css, name)` | string or `None` | Attribute from first matching descendant of `doc` or `node`. |
| `node_name(node)` | string | HTML node name such as `h2` or `span`. |
| `has_parent(node, css)` | bool | Whether a node has a matching ancestor. |
| `clone_without_text(node, css)` | string | Clone node, remove matching descendants, return text. |
| `trim(s)` | string | Trim leading and trailing whitespace. |
| `collapse_space(s)` | string | Replace all whitespace runs with one space. |
| `url_join(base, ref)` | string | Resolve `ref` against `base`. |
| `url_without_anchor(url)` | string | Remove URL fragment. |
| `path(url)` | string | URL path only. |
| `sha1(s)` | string | Hex SHA-1 digest for stable IDs. |
| `regex_match(pattern, s)` | bool | RE2 regular expression match. |
| `regex_replace(pattern, repl, s)` | string | RE2 regular expression replacement. |

Example minimal script: `examples/minimal.star`.

Validate a script without crawling:

```sh
docs-crawler script check --script examples/algolia.star
```

Emit script info as JSON for tooling:

```sh
docs-crawler script check --script examples/algolia.star --json
```

Debug extractor matching and record counts during crawls:

```sh
docs-crawler crawl single --script examples/minimal.star --debug-script https://example.com/doc
```

## Output

Output is newline-delimited JSON (`.jsonl`). Each line is one script-produced record returned by the matching extractor.

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

## Library API

Core interfaces live in `pkg/crawler`:

```go
type Source interface {
    URLs(ctx context.Context) ([]string, error)
}

type Fetcher interface {
    Fetch(ctx context.Context, ref string) (model.Page, error)
}

type Parser interface {
    Parse(page model.Page) (model.ParsedPage, error)
}

type Extractor interface {
    Extract(ctx context.Context, page model.ParsedPage) ([]any, error)
}

type Writer interface {
    Write(record any) error
    Close() error
}
```

Run a pipeline with the built-in HTTP/HTML/Starlark pieces:

```go
extractor, err := extract.NewStarlarkExtractor("examples/algolia.star", false)
if err != nil {
    return err
}

err = crawler.Run(ctx, crawler.Pipeline{
    Source:    source.Sitemap{SitemapURL: "https://www.algolia.com/doc/sitemap.xml"},
    Fetcher:   fetch.HTTPFetcher{},
    Parser:    parse.HTMLParser{},
    Extractor: extractor,
    Writer:    output.NewJSONLWriter(os.Stdout),
    Workers:   8,
})
```

`model.Page` is raw fetched content:

```go
type Page struct {
    Ref         string         // original reference: URL, path, key, etc.
    URL         string         // canonical URL when available
    StatusCode  int            // HTTP status when available
    ContentType string         // content type or equivalent hint
    Body        []byte
    Metadata    map[string]any
}
```

`model.ParsedPage` is parser output:

```go
type ParsedPage struct {
    Ref      string
    URL      string
    Kind     string // "html", "markdown", etc.
    Document any    // parser-specific document
    Metadata map[string]any
}
```

### Custom pipeline example

This is the intended extension point: implement only the stages you need.

```go
type Files struct { Paths []string }

func (s Files) URLs(context.Context) ([]string, error) {
    return s.Paths, nil
}

type FileFetcher struct{}

func (FileFetcher) Fetch(_ context.Context, path string) (model.Page, error) {
    body, err := os.ReadFile(path)
    if err != nil {
        return model.Page{}, err
    }
    return model.Page{Ref: path, Body: body, ContentType: "text/markdown"}, nil
}

type MarkdownParser struct{}

func (MarkdownParser) Parse(page model.Page) (model.ParsedPage, error) {
    doc := parseMarkdown(page.Body) // your code
    return model.ParsedPage{Ref: page.Ref, Kind: "markdown", Document: doc}, nil
}

type MarkdownExtractor struct{}

func (MarkdownExtractor) Extract(_ context.Context, page model.ParsedPage) ([]any, error) {
    doc := page.Document.(*MarkdownDoc) // your type
    return []any{{"path": page.Ref, "title": doc.Title}}, nil
}

err := crawler.Run(ctx, crawler.Pipeline{
    Source:    Files{Paths: []string{"docs/intro.md"}},
    Fetcher:   FileFetcher{},
    Parser:    MarkdownParser{},
    Extractor: MarkdownExtractor{},
    Writer:    output.NewJSONLWriter(os.Stdout),
})
```

Use built-in packages when they fit:

- `pkg/source`: `Single`, `Sitemap`
- `pkg/fetch`: `HTTPFetcher`
- `pkg/parse`: `HTMLParser`
- `pkg/extract`: `StarlarkExtractor`
- `pkg/output`: `JSONLWriter`

## Project layout

- `main.go`: CLI entrypoint
- `cmd`: CLI commands and flags
- `pkg/crawler`: public crawl pipeline orchestration and stage interfaces
- `pkg/source`: URL discovery for single URL or sitemap
- `pkg/fetch`: HTTP page fetching
- `pkg/parse`: HTML parsing
- `pkg/extract`: extractor implementations
- `pkg/script`: language-neutral script interfaces and JSON validation
- `pkg/script/starlark`: Starlark engine and host API
- `pkg/output`: JSONL writer
- `examples/algolia.star`: Algolia docs extractor DSL script

## Test

```bash
go test ./...
```

Full validation:

```bash
mise all
```
