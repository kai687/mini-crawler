package source

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestLLMSURLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != "mini-crawler/0.1" {
			t.Fatalf("User-Agent = %q", r.UserAgent())
		}

		_, _ = w.Write([]byte(`# Docs

- [Intro](/docs/intro.md)
- [Guide](guide.md)
- [External](https://example.org/api.md)
- [Duplicate](/docs/intro.md)
- [Anchor](#overview)
- [Email](mailto:docs@example.com)
`))
	}))
	defer server.Close()

	l := LLMS{
		LLMSURL: server.URL + "/llms.txt",
		Client:  server.Client(),
	}

	urls, err := l.URLs(context.Background())
	if err != nil {
		t.Fatalf("URLs() err = %v", err)
	}

	want := []string{
		server.URL + "/docs/intro.md",
		server.URL + "/guide.md",
		"https://example.org/api.md",
	}
	if !reflect.DeepEqual(urls, want) {
		t.Fatalf("URLs() = %#v, want %#v", urls, want)
	}
}

func TestLLMSURLsUseFinalRedirectURLAsBase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/old/llms.txt":
			http.Redirect(w, r, "/new/llms.txt", http.StatusFound)
		case "/new/llms.txt":
			_, _ = w.Write([]byte(`[Guide](guide.md)`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	urls, err := LLMS{LLMSURL: server.URL + "/old/llms.txt", Client: server.Client()}.
		URLs(context.Background())
	if err != nil {
		t.Fatalf("URLs() err = %v", err)
	}

	want := []string{server.URL + "/new/guide.md"}
	if !reflect.DeepEqual(urls, want) {
		t.Fatalf("URLs() = %#v, want %#v", urls, want)
	}
}

func TestParseLLMSAllowsMarkdownLinkTitle(t *testing.T) {
	base := mustParseURL(t, "https://example.com/docs/llms.txt")

	urls, err := parseLLMS(strings.NewReader(`[Intro](/intro.md "Intro page")`), base)
	if err != nil {
		t.Fatalf("parseLLMS() err = %v", err)
	}

	want := []string{"https://example.com/intro.md"}
	if !reflect.DeepEqual(urls, want) {
		t.Fatalf("parseLLMS() = %#v, want %#v", urls, want)
	}
}

func TestResolveLLMSURLSkipsUnsupportedSchemes(t *testing.T) {
	base := mustParseURL(t, "https://example.com/llms.txt")

	for _, raw := range []string{"#intro", "mailto:docs@example.com", "javascript:alert(1)"} {
		got, ok, err := resolveLLMSURL(base, raw)
		if err != nil {
			t.Fatalf("resolveLLMSURL(%q) err = %v", raw, err)
		}

		if ok || got != "" {
			t.Fatalf("resolveLLMSURL(%q) = %q, %v; want skipped", raw, got, ok)
		}
	}
}
