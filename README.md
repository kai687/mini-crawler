# mini-crawler

Small CLI for extracting JSON records from HTML pages.
You can then index these records into search engines or databases.

## How it works

The CLI discovers URLs from a sitemap and extracts information from each HTML page.
You can control what gets extracted by scripts.

## Requirements

**Only crawl sites you own.**
Your site must have a sitemap for URL discovery.
This crawler does not crawl links found in documents.

## Usage

Run `mini-crawler --help` to see the available commands and options.

When developing your extraction logic,
it can be useful to extract only a single URL:

```sh
mini-crawler crawl single URL --script EXTRACTION.STAR
```

By default, extracted records are printed to standard output.
To write them into a file, use the `--output` option.

To crawl your site, run:

```sh
mini-crawler crawl sitemap SITEMAP_URL --script EXTRACTION.STAR --workers 8 --output records.jsonl
```

This processes your sitemap in parallel.
If you plan to use this program regularly,
run tests to see how much you actually benefit from parallelism.

## Extraction scripts

Extraction is controlled by scripts written in [Starlark](https://github.com/google/starlark-go),
a Python-like scripting language.

Scripts register extractor functions with the `extract(pattern, fn)` DSL:

```python
def handle_guides(pattern, doc, ctx):
    return [{"url": ctx["url"], "title": text(doc.select_first("h1"))}]


extract("^/doc/guides/", handle_guides)
```

Rules for extractor functions:

- Extractor functions must accept exactly three arguments: `pattern`, `doc`, `ctx`
- `pattern` must be a valid regular expression matched against the URL path
- Registration order matters; first matching extractor wins
- Extractor functions must return a list of JSON-like records
- If no extractor matches, the page is skipped with a warning

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

Output is newline-delimited JSON (`.jsonl`).


## Library usage

You can use this program as a library.
A crawl has five replaceable stages:

1. **Source** discovers references to process. A reference can be an HTTP URL, file path, object key, database ID, or any string your fetcher understands.
1. **Fetcher** loads raw bytes for one reference.
1. **Parser** turns raw bytes into a parsed document shape.
1. **Extractor** reads the parsed document and returns JSON-like records.
1. **Writer** receives records and persists them.

You can swap any stage, for example, for crawling local Markdown files:

```text
filesystem source -> file fetcher -> Markdown parser -> custom extractor -> JSONL writer
```

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

## Development

### Build

```sh
mise build
```

### Tests

```sh
mise test
```

### Lint and format

Build the binary, lint, format, and test:

```sh
mise all
```
