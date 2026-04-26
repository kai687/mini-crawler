# docs-crawler

Small Go CLI for crawling documentation pages and emitting Algolia-ready records as JSONL.

Tool currently tuned for Algolia docs HTML structure and Algolia search settings. Output schema and record splitting are designed to work with current `docs_clean` index ranking, not as a generic docs crawler format.

It supports two input modes:
- `sitemap`: load page URLs from a sitemap, then crawl each page
- `single`: crawl one explicit page URL

For each page, crawler:
1. fetches HTML
2. parses document
3. extracts page title, headings, content, and API field blocks
4. enriches data with breadcrumbs, hierarchy, content type, and stable object IDs
5. writes one JSON object per record

## Requirements

- Go `1.26+`

## Build

```bash
go build -o docs-crawler .
```

## Usage

```bash
docs-crawler [flags] <sitemap-url>
docs-crawler --mode single [flags] <url>
```

### Flags

- `--mode`: crawl mode, `sitemap` or `single`. Default: `sitemap`
- `--workers`: number of concurrent page workers. Default: `1`
- `--filter`: substring filter applied to sitemap URLs before crawling
- `--fail-on-error`: stop run if one URL fails
- `--coverage`: print crawl coverage summary to stderr
- `--verbose`: show crawl logs on stderr
- `--output`: write JSONL to file instead of stdout

## Examples

Crawl one page and print records to stdout:

```bash
go run . --mode single https://algolia.com/doc/ui-libraries/autocomplete/introduction/what-is-autocomplete
```

Crawl sitemap with 8 workers and save output:

```bash
go run . \
  --workers 8 \
  --output records.jsonl \
  https://algolia.com/sitemap.xml
```

Only crawl sitemap URLs containing `/rest-api/`:

```bash
go run . \
  --filter /rest-api/ \
  --coverage \
  --output records.jsonl \
  https://algolia.com/sitemap.xml
```

Fail fast on first page error:

```bash
go run . --workers 4 --fail-on-error https://algolia.com/sitemap.xml
```

## Hard-coded assumptions and caveats

This crawler is not site-agnostic. Several behaviors are hard coded around current Algolia docs structure.

### HTML structure assumptions

Crawler only indexes content under first matching root:
- `#content`
- `#content-area`

Inside content root, it only extracts:
- headings: `h2[id]`, `h3[id]`, `h4[id]`, `h5[id]`, `h6[id]`
- paragraph-like text: `span[data-as='p']`
- list items: `li` with non-empty text after links are stripped
- API field headers: `div.param-head[id]`

Page-level metadata is also hard coded:
- page title from `h1#page-title`
- page description from `meta[name='description']`

If site markup changes, crawler may silently miss content or produce lower heading coverage.

### URL-derived assumptions

Several fields come from URL shape, not page DOM:
- breadcrumb labels are inferred from URL path segments
- leading `/doc` is stripped when building breadcrumb path
- final path segment is excluded from breadcrumbs because current page label comes from page title / heading
- `contentType` is inferred from path prefix only:
  - `/guides` -> `guide`
  - `/rest-api` -> `api`
  - `/api-reference/api-parameters` -> `api`
  - `/integration` -> `integration`
  - `/libraries/sdk` -> `sdk`
  - `/framework-integration` -> `sdk`
  - anything else -> empty string
- `variant` is inferred from URL shape only:
  - `/libraries/sdk/v1` -> `legacy`
  - `/rest-api/ingestion/*v1` -> `legacy`
  - anything else -> empty string

### Record model assumptions

Crawler always emits one page-level `lvl1` record first, even if page has no extracted body content.

API field records are forced into hierarchy level 3 semantics:
- `recordType = field`
- field name stored in `hierarchy.lvl3`
- field metadata/description stored in `content`

This works for current API reference structure, but is opinionated.

### Search/indexing assumptions

Output is designed for current Algolia index settings on `docs_clean`.

In particular:
- `urlWithoutAnchor` exists because index uses `distinct=true` with `attributeForDistinct=urlWithoutAnchor`
- `position` exists because index uses `customRanking=["asc(position)"]`
- `hierarchy.lvl0` ... `hierarchy.lvl6` and `content` exist because they are current `searchableAttributes`
- breadcrumbs and `contentType` are currently metadata/faceting/display helpers; current fetched settings facet only on `contentType`

If ranking/settings change, ideal record shape may also need to change.

## Output

Output format is newline-delimited JSON (`.jsonl`). Each line is one record.

Record types currently include:
- `lvl1`: page-level record
- `lvl2` ... `lvl6`: heading records
- `field`: API field records
- `content`: paragraph or list-item content

Example line:

```json
{"url":"https://algolia.com/doc/rest-api/search#query","urlWithoutAnchor":"https://algolia.com/doc/rest-api/search","breadcrumbSegments":["REST API"],"breadcrumbHierarchy":{"lvl0":"REST API"},"contentType":"api","recordType":"field","content":"string. required. Search query text.","hierarchy":{"lvl1":"Search","lvl3":"query"},"position":12,"objectID":"doc-rest-api-search-query"}
```

## Record structure

Each JSON line is one `Record`.

### Field summary table

| Field | Type | How populated | Why it exists |
|---|---|---|---|
| `url` | string | Page URL, or page URL + `#anchor` for heading/field records. Content records use current active heading/field URL. | Destination URL for hit. Lets different records from same page point to exact section. |
| `urlWithoutAnchor` | string | `url` with fragment removed. | Distinct key. Current settings use `attributeForDistinct="urlWithoutAnchor"`. |
| `breadcrumbSegments` | `string[]` | Humanized URL path segments, excluding final path segment. | Display/context metadata. |
| `breadcrumbHierarchy` | object | Cumulative hierarchy built from `breadcrumbSegments`. | Hierarchical metadata for UI/faceting if needed. |
| `contentType` | string | Inferred from URL prefix: `/guides` -> `guide`, `/rest-api` -> `api`, `/libraries/sdk` -> `sdk`. | Current faceting/filter field. |
| `variant` | string | Inferred from URL shape: `/libraries/sdk/v1` -> `legacy`; `/rest-api/ingestion/*v1` -> `legacy`. Omitted otherwise. | Extra filter/facet field for legacy doc subsets. |
| `methodName` | string | For `/doc/rest-api/...` and `/doc/libraries/sdk/methods/...` URLs, camelCase from final path slug, for example `/search-single-index` -> `searchSingleIndex`. | Exact API method-name recall + boost field. |
| `recordType` | string | Derived from extracted unit kind and heading depth. | Tells what semantic unit record represents. |
| `content` | string/null | Page description, body text, or synthesized field description depending on record type. | Main full-text body field. Also snippet source. |
| `hierarchy` | object | Page title in `lvl1`, active heading stack in `lvl2`-`lvl6`. | Main title/heading search fields. |
| `position` | integer | `0` for page record, then DOM-order index for extracted units. | Current custom ranking key: earlier content ranks higher. |
| `objectID` | string | Stable ID from normalized URL; content records append `-<position>`. | Unique Algolia record ID. |

### Record type table

| `recordType` | What it represents | `url` behavior | `content` behavior | `hierarchy` behavior |
|---|---|---|---|---|
| `lvl1` | Page-level record | Bare page URL | Page meta description, if present | `lvl1` = page title |
| `lvl2` | `h2[id]` heading | URL + heading anchor | Omitted | `lvl1` + `lvl2` set |
| `lvl3` | `h3[id]` heading | URL + heading anchor | Omitted | `lvl1` + active heading stack through `lvl3` |
| `lvl4` | `h4[id]` heading | URL + heading anchor | Omitted | `lvl1` + active heading stack through `lvl4` |
| `lvl5` | `h5[id]` heading | URL + heading anchor | Omitted | `lvl1` + active heading stack through `lvl5` |
| `lvl6` | `h6[id]` heading | URL + heading anchor | Omitted | `lvl1` + active heading stack through `lvl6` |
| `field` | API field / parameter / property block | URL + field anchor | Synthesized from type pill, required pill, first description paragraph | Field name forced into `hierarchy.lvl3` |
| `content` | Paragraph or list-item text | Current active heading/field URL | Raw paragraph/list-item text | Inherits current heading stack |

### Field details

#### `url`
Canonical URL for this record.

Meaning depends on record type:
- page record (`lvl1`): page URL without fragment
- heading record (`lvl2` ... `lvl6`): page URL with heading `#anchor`
- field record (`field`): page URL with field block `#anchor`
- content record (`content`): URL of most recent heading/field context, or bare page URL if content appears before first heading

Important because:
- destination URL used by search hit
- multiple records from same page can point to different anchors

#### `urlWithoutAnchor`
Page URL with fragment removed.

Important because:
- current Algolia settings use `attributeForDistinct: "urlWithoutAnchor"`
- groups hits from same page for `distinct=true`
- search can rank many records from one page internally, then show one representative page hit

#### `breadcrumbSegments`
Human-readable breadcrumb labels derived from URL path segments, excluding final path segment.

Example:
- URL: `/doc/ui-libraries/autocomplete/guides/debugging`
- `breadcrumbSegments`: `["UI libraries", "Autocomplete", "Guides"]`

#### `breadcrumbHierarchy`
Cumulative breadcrumb levels derived from `breadcrumbSegments`.

Example for `["UI libraries", "Autocomplete", "Guides"]`:

```json
{
  "lvl0": "UI libraries",
  "lvl1": "UI libraries > Autocomplete",
  "lvl2": "UI libraries > Autocomplete > Guides"
}
```

Current fetched `docs_clean` settings do not facet on these fields, but schema keeps them for search UI or future faceting.

#### `contentType`
High-level category inferred from URL path.

Current values:
- `guide` for paths starting with `/guides`
- `api` for paths starting with `/rest-api`
- `sdk` for paths starting with `/libraries/sdk`
- omitted otherwise

Important because current Algolia settings facet on `contentType`.

#### `variant`
URL-path variant classification.

Current values:
- `legacy` for paths starting with `/libraries/sdk/v1`
- `legacy` for `/rest-api/ingestion/` paths whose final segment ends with `v1` (for example `/update-task-v1`)
- omitted otherwise

Useful for filtering legacy SDK docs separately from current SDK docs.

#### `recordType`
Kind of record.

Values:
- `lvl1`: page-level record
- `lvl2` ... `lvl6`: heading records by depth
- `field`: API field/parameter/property record
- `content`: paragraph or list-item body content

Useful for debugging extraction and ranking behavior.

#### `content`
Body text associated with record.

Population rules:
- page record (`lvl1`): page meta description, if present
- heading record (`lvl2` ... `lvl6`): omitted
- field record (`field`): synthesized field description, built from type pill, required pill, and first description paragraph
- content record (`content`): paragraph or list-item text

Important because:
- current index has `unordered(content)` in `searchableAttributes`
- this is main long-tail recall field
- current settings snippet it with `attributesToSnippet=["content:20"]`

#### `hierarchy`
Heading context for record.

Shape:

```json
{
  "lvl0": "...",
  "lvl1": "...",
  "lvl2": "...",
  "lvl3": "...",
  "lvl4": "...",
  "lvl5": "...",
  "lvl6": "..."
}
```

Current population rules:
- `hierarchy.lvl1`: page title from `h1#page-title`
- `hierarchy.lvl2` ... `hierarchy.lvl6`: current active heading stack while scanning page in DOM order
- lower levels are cleared when crawler enters shallower heading
- `lvl0` exists in schema but crawler does not populate it yet
- field records force field name into `hierarchy.lvl3`

Important because:
- current index ranking starts with `attribute`
- current `searchableAttributes` put `hierarchy.*` before `content`
- title/heading matches therefore generally outrank body-text matches
- record splitting + hierarchy fields let search match exact section, not just whole page

#### `position`
Extraction order within page.

Rules:
- page record always has `position = 0`
- extracted units then increment in DOM order starting at `1`
- content records keep observed order relative to headings/fields

Important because current index uses `customRanking=["asc(position)"]`, so earlier content beats later content when higher ranking criteria tie.

#### `objectID`
Stable identifier for Algolia record.

Construction rules:
- based on normalized path/fragment/query, not full scheme+host when URL parses cleanly
- slashes replaced with `-`
- content records append `-<position>` because many content records share same anchored URL

Examples:
- page record: `doc-rest-api-search`
- heading record: `doc-rest-api-search#query`
- content record under same anchor at position 12: `doc-rest-api-search#query-12`

Important because:
- required unique primary key for indexing
- deterministic IDs make re-indexing idempotent

## How record types work together with ranking

Current `docs_clean` settings:

```json
{
  "attributeForDistinct": "urlWithoutAnchor",
  "attributesForFaceting": ["contentType"],
  "attributesToSnippet": ["content:20"],
  "customRanking": ["asc(position)"],
  "distinct": true,
  "ranking": ["attribute", "typo", "words", "filters", "proximity", "exact", "custom"],
  "searchableAttributes": [
    "unordered(methodName)",
    "unordered(hierarchy.lvl0)",
    "unordered(hierarchy.lvl1)",
    "unordered(hierarchy.lvl2)",
    "unordered(hierarchy.lvl3)",
    "unordered(hierarchy.lvl4)",
    "unordered(hierarchy.lvl5)",
    "unordered(hierarchy.lvl6)",
    "unordered(content)"
  ]
}
```

Practical effect:

1. crawler emits many records per page
   - page title record
   - heading records
   - field records
   - content records

2. Algolia first prefers matches in earlier searchable attributes
   - `methodName` beats title/heading/body when query matches canonical API method name
   - title/heading match beats body-content match
   - shallower hierarchy field can beat deeper/body fields when rest is equal

3. within similar matches, earlier page position wins
   - because `customRanking` is `asc(position)`

4. final result set is deduplicated per page
   - because `distinct=true`
   - distinct key is `urlWithoutAnchor`

So record model intentionally balances two goals that would conflict in a single-record-per-page design:
- fine-grained matching at section/field/content level
- one final visible hit per page

Without record splitting, search would lose section precision.
Without `distinct`, users could get many near-duplicate hits from same page.
Without `position`, late-page incidental text could outrank early defining text too often.

## Extraction rules

Crawler looks for content under first matching root:
- `#content`
- `#content-area`

Inside content root, it extracts:
- headings: `h2[id]` through `h6[id]`
- paragraph-like text: `span[data-as='p']`
- list items: `li` (ignoring anchor-only items)
- API field headers: `div.param-head[id]`

For API field headers, crawler synthesizes field description from:
- field type pill
- required pill
- first paragraph in next `div.mt-4` description block

Page-level metadata comes from:
- title: `h1#page-title`
- description: `meta[name='description']`

## Coverage summary

With `--coverage`, tool prints summary to stderr after crawl:

- total URLs seen
- URLs that produced records
- total record count
- average records per URL
- extracted heading coverage vs expected headings in content root

Example:

```text
coverage: 120/123 URLs returned records (97.56%); 8450 records total; average 68.70 records/URL; headings 932/950 (98.11%)
```

Use this as regression signal when markup changes. Big heading-coverage drop often means selector drift.

## Project layout

- `main.go`: CLI entrypoint and output selection
- `internal/config`: flag parsing and validation
- `internal/source`: URL discovery for single URL or sitemap
- `internal/fetch`: HTTP page fetching
- `internal/parse`: HTML parsing
- `internal/extract`: raw content extraction from parsed pages
- `internal/enrich`: transform extracted units into final records
- `internal/output`: JSONL writer
- `internal/coverage`: crawl coverage metrics
- `internal/recordutil`: URL, breadcrumb, and object ID helpers

## Test

```bash
go test ./...
```
