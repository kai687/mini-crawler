package script

// Engine loads site-specific extraction programs from disk.
type Engine interface {
	Load(path string) (Program, error)
}

// Program extracts and enriches JSON-like records for one parsed page.
type Program interface {
	PageMeta(doc Document, ctx Context) (map[string]any, error)
	Records(doc Document, ctx Context) ([]map[string]any, error)
	Enrich(record map[string]any, ctx Context) (map[string]any, error)
}
