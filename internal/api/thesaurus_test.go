package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const thesFixture = `[
  {
    "meta": {
      "id": "happy",
      "syns": [["glad", "joyful", "blissful", "cheerful"]],
      "ants": [["sad", "unhappy", "depressed", "miserable"]],
      "offensive": false
    },
    "hwi": {"hw": "hap*py"},
    "fl": "adjective",
    "def": []
  }
]`

const thesMultiSenseFixture = `[
  {
    "meta": {
      "id": "fast:1",
      "syns": [["quick", "speedy", "swift"], ["fixed", "secure", "firm"]],
      "ants": [["slow", "sluggish"], ["loose", "unfixed"]],
      "offensive": false
    },
    "hwi": {"hw": "fast"},
    "fl": "adjective",
    "def": []
  }
]`

func TestThesaurus_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(thesFixture))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	entries, err := c.Thesaurus("happy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Fl != "adjective" {
		t.Errorf("Fl = %q, want \"adjective\"", e.Fl)
	}
	if len(e.Meta.Syns) != 1 {
		t.Errorf("expected 1 syn group, got %d", len(e.Meta.Syns))
	}
	if len(e.Meta.Syns[0]) != 4 {
		t.Errorf("expected 4 synonyms, got %d", len(e.Meta.Syns[0]))
	}
	if len(e.Meta.Ants) != 1 {
		t.Errorf("expected 1 ant group, got %d", len(e.Meta.Ants))
	}
}

func TestThesaurus_MultiSense(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(thesMultiSenseFixture))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	entries, err := c.Thesaurus("fast")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries[0].Meta.Syns) != 2 {
		t.Errorf("expected 2 syn sense groups, got %d", len(entries[0].Meta.Syns))
	}
	if len(entries[0].Meta.Ants) != 2 {
		t.Errorf("expected 2 ant sense groups, got %d", len(entries[0].Meta.Ants))
	}
}

func TestThesaurus_WordNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`["happy","happiness","happily"]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Thesaurus("hppy")
	if _, ok := err.(*WordNotFoundError); !ok {
		t.Errorf("expected *WordNotFoundError, got %T: %v", err, err)
	}
}

func TestThesaurus_UsesThesaurusKey(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.apiKeyThesaurus = "thes-secret"
	c.Thesaurus("happy")

	if gotQuery != "key=thes-secret" {
		t.Errorf("query = %q, want \"key=thes-secret\"", gotQuery)
	}
}
