package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient returns a Client pointed at the given httptest server.
func newTestClient(server *httptest.Server) *Client {
	return &Client{
		http:            server.Client(),
		apiKeyDict:      "dict-key",
		apiKeyThesaurus: "thes-key",
		baseURL:         server.URL,
	}
}

func TestGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return a minimal array of objects (not strings) — simulates a found word.
		w.Write([]byte(`[{"meta":{"id":"run:1","stems":["run"],"offensive":false},"hwi":{"hw":"run"},"fl":"verb","def":[],"et":[],"date":""}]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	var entries []DictEntry
	err := c.get("collegiate", "run", "dict-key", &entries)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Fl != "verb" {
		t.Errorf("expected fl=verb, got %q", entries[0].Fl)
	}
}

func TestGet_WordNotFound_WithSuggestions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`["running","runner","runway"]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	var entries []DictEntry
	err := c.get("collegiate", "runn", "dict-key", &entries)
	if err == nil {
		t.Fatal("expected WordNotFoundError, got nil")
	}
	wnf, ok := err.(*WordNotFoundError)
	if !ok {
		t.Fatalf("expected *WordNotFoundError, got %T: %v", err, err)
	}
	if wnf.Word != "runn" {
		t.Errorf("expected Word=runn, got %q", wnf.Word)
	}
	if len(wnf.Suggestions) != 3 {
		t.Errorf("expected 3 suggestions, got %d", len(wnf.Suggestions))
	}
}

func TestGet_WordNotFound_EmptyArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	var entries []DictEntry
	err := c.get("collegiate", "xyzzy", "dict-key", &entries)
	if err == nil {
		t.Fatal("expected error for empty array")
	}
	if _, ok := err.(*WordNotFoundError); !ok {
		t.Fatalf("expected *WordNotFoundError, got %T", err)
	}
}

func TestGet_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid API key.", http.StatusForbidden)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	var entries []DictEntry
	err := c.get("collegiate", "run", "bad-key", &entries)
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected error to mention HTTP 403, got: %v", err)
	}
}

func TestGet_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Invalid API key.`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	var entries []DictEntry
	err := c.get("collegiate", "run", "dict-key", &entries)
	if err == nil {
		t.Fatal("expected error for non-JSON response")
	}
	if !strings.Contains(err.Error(), "Invalid API key.") {
		t.Errorf("error should contain the raw API message, got: %v", err)
	}
}

func TestGet_URLConstruction(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.String()
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	var entries []DictEntry
	c.get("collegiate", "serendipity", "mykey", &entries)

	want := "/collegiate/json/serendipity?key=mykey"
	if gotPath != want {
		t.Errorf("URL path = %q, want %q", gotPath, want)
	}
}
