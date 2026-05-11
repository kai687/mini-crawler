package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/algolia/mini-crawler/pkg/httpheaders"
	"github.com/algolia/mini-crawler/pkg/model"
)

// HTTPFetcher retrieves pages with an http.Client.
//
// It returns raw bytes only; HTML interpretation stays in the parse package.
type HTTPFetcher struct {
	// Client performs HTTP requests. Nil uses http.DefaultClient.
	Client *http.Client
}

// Fetch downloads one page URL and returns raw response data.
func (f HTTPFetcher) Fetch(ctx context.Context, pageURL string) (model.Page, error) {
	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return model.Page{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", httpheaders.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return model.Page{}, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Page{}, fmt.Errorf("read body: %w", err)
	}

	return model.Page{
		Ref:         pageURL,
		URL:         pageURL,
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Body:        body,
	}, nil
}
