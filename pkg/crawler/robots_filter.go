package crawler

import (
	"context"
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/mini-crawler/pkg/model"
)

// ErrNoindex means parsed page declares robots noindex and should be skipped.
var ErrNoindex = errors.New("robots noindex")

// RobotsNoindexFilter skips HTML pages with robots noindex metadata.
type RobotsNoindexFilter struct{}

func (RobotsNoindexFilter) FilterParsedPage(_ context.Context, page model.ParsedPage) error {
	if hasRobotsNoindex(page) {
		return errors.Join(ErrFiltered, ErrNoindex)
	}

	return nil
}

func hasRobotsNoindex(page model.ParsedPage) bool {
	if page.Doc == nil {
		return false
	}

	found := false

	page.Doc.Find("meta").EachWithBreak(func(_ int, meta *goquery.Selection) bool {
		content, ok := robotsMetaContent(meta)
		if !ok {
			return true
		}

		found = directivesContainNoindex(content)

		return !found
	})

	return found
}

func robotsMetaContent(meta *goquery.Selection) (string, bool) {
	name, ok := meta.Attr("name")
	if !ok || !strings.EqualFold(strings.TrimSpace(name), "robots") {
		return "", false
	}

	return meta.Attr("content")
}

func directivesContainNoindex(content string) bool {
	for _, directive := range strings.FieldsFunc(content, isRobotsDirectiveSeparator) {
		if strings.EqualFold(strings.TrimSpace(directive), "noindex") {
			return true
		}
	}

	return false
}

func isRobotsDirectiveSeparator(r rune) bool {
	return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
