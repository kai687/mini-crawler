package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPFetcherFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != "mini-crawler/0.1" {
			t.Fatalf("User-Agent = %q", r.UserAgent())
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer server.Close()

	page, err := HTTPFetcher{Client: server.Client()}.Fetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Fetch() err = %v", err)
	}

	if page.URL != server.URL {
		t.Fatalf("URL = %q, want %q", page.URL, server.URL)
	}

	if page.StatusCode != http.StatusAccepted {
		t.Fatalf("StatusCode = %d, want %d", page.StatusCode, http.StatusAccepted)
	}

	if page.ContentType != "text/html; charset=utf-8" {
		t.Fatalf("ContentType = %q", page.ContentType)
	}

	if string(page.Body) != "<html><body>ok</body></html>" {
		t.Fatalf("Body = %q", string(page.Body))
	}
}

func TestHTTPFetcherUsesFinalRedirectURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/old":
			http.Redirect(w, r, "/new", http.StatusFound)
		case "/new":
			_, _ = w.Write([]byte("ok"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	page, err := HTTPFetcher{Client: server.Client()}.Fetch(context.Background(), server.URL+"/old")
	if err != nil {
		t.Fatalf("Fetch() err = %v", err)
	}

	want := server.URL + "/new"
	if page.URL != want {
		t.Fatalf("URL = %q, want %q", page.URL, want)
	}

	if page.Ref != server.URL+"/old" {
		t.Fatalf("Ref = %q, want original URL", page.Ref)
	}
}

func TestHTTPFetcherFetchError(t *testing.T) {
	_, err := HTTPFetcher{Client: &http.Client{}}.Fetch(context.Background(), "://bad-url")
	if err == nil {
		t.Fatal("Fetch() err = nil, want error")
	}

	if !strings.Contains(err.Error(), "build request") {
		t.Fatalf("err = %v, want build request error", err)
	}
}
