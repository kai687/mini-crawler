package script

// Engine loads site-specific extraction programs from disk.
type Engine interface {
	Load(path string) (Program, error)
}

// Program extracts JSON-like records for one parsed page.
type Program interface {
	Extract(doc Document, ctx Context) ([]map[string]any, error)
}
