package recordutil

// URLWithAnchor returns pageURL with a fragment anchor when anchor is non-empty.
func URLWithAnchor(pageURL string, anchor string) string {
	if anchor == "" {
		return pageURL
	}

	return pageURL + "#" + anchor
}
