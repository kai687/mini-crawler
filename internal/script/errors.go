package script

import "errors"

// ErrNoExtractor means no registered extractor matched a page URL.
var ErrNoExtractor = errors.New("no extractor matches")
