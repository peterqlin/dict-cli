package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Realistic fixture trimmed from the actual MW Collegiate Dictionary response.
const dictFixture = `[
  {
    "meta": {"id": "serendipity:1", "stems": ["serendipity"], "offensive": false},
    "hwi": {
      "hw": "ser*en*dip*i*ty",
      "prs": [{"mw": "ˌser-ən-ˈdi-pə-tē"}]
    },
    "fl": "noun",
    "def": [
      {
        "sseq": [
          [
            ["sense", {
              "sn": "1",
              "dt": [["text", "the faculty or phenomenon of finding valuable or agreeable things not sought for"]]
            }]
          ]
        ]
      }
    ],
    "et": [["text", "coined by Horace Walpole"]],
    "date": "1754"
  }
]`

func TestDefine_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(dictFixture))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	entries, err := c.Define("serendipity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Fl != "noun" {
		t.Errorf("Fl = %q, want \"noun\"", e.Fl)
	}
	if e.Hwi.Hw != "ser*en*dip*i*ty" {
		t.Errorf("Hwi.Hw = %q, want \"ser*en*dip*i*ty\"", e.Hwi.Hw)
	}
	if len(e.Hwi.Prs) == 0 {
		t.Error("expected at least one pronunciation")
	}
	if e.Date != "1754" {
		t.Errorf("Date = %q, want \"1754\"", e.Date)
	}
}

func TestDefine_WordNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`["serendipity","serenity","serendipitous"]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Define("serndipity")
	if err == nil {
		t.Fatal("expected error")
	}
	wnf, ok := err.(*WordNotFoundError)
	if !ok {
		t.Fatalf("expected *WordNotFoundError, got %T", err)
	}
	if len(wnf.Suggestions) != 3 {
		t.Errorf("expected 3 suggestions, got %d", len(wnf.Suggestions))
	}
}

func TestDefine_MultipleEntries(t *testing.T) {
	// "run" has many homographs in MW.
	fixture := `[
		{"meta":{"id":"run:1","stems":["run"],"offensive":false},"hwi":{"hw":"run"},"fl":"verb","def":[],"et":[],"date":""},
		{"meta":{"id":"run:2","stems":["run"],"offensive":false},"hwi":{"hw":"run"},"fl":"noun","def":[],"et":[],"date":""}
	]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fixture))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	entries, err := c.Define("run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}
