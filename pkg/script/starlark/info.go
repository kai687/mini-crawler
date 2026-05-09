package starlark

// ExtractorInfo describes one extractor registered by a script.
type ExtractorInfo struct {
	Pattern string `json:"pattern"`
	Name    string `json:"name"`
}

// Info describes a loaded Starlark script.
type Info struct {
	Path       string          `json:"path"`
	Extractors []ExtractorInfo `json:"extractors"`
}

// Check loads one Starlark script and returns registered extractor details.
func Check(path string) (Info, error) {
	program, err := Engine{}.Load(path)
	if err != nil {
		return Info{}, err
	}

	starlarkProgram := program.(*Program)

	info := Info{
		Path:       path,
		Extractors: make([]ExtractorInfo, 0, len(starlarkProgram.extractors)),
	}
	for _, extractor := range starlarkProgram.extractors {
		info.Extractors = append(info.Extractors, ExtractorInfo{
			Pattern: extractor.pattern,
			Name:    extractor.fn.Name(),
		})
	}

	return info, nil
}
