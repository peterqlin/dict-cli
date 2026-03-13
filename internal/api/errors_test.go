package api

import (
	"strings"
	"testing"
)

func TestWordNotFoundError_WithSuggestions(t *testing.T) {
	err := &WordNotFoundError{Word: "runn", Suggestions: []string{"run", "runner", "running"}}
	msg := err.Error()
	if !strings.Contains(msg, "runn") {
		t.Errorf("error message should contain the word, got: %q", msg)
	}
	if !strings.Contains(msg, "did you mean") {
		t.Errorf("error message should mention suggestions, got: %q", msg)
	}
	if !strings.Contains(msg, "runner") {
		t.Errorf("error message should list suggestions, got: %q", msg)
	}
}

func TestWordNotFoundError_NoSuggestions(t *testing.T) {
	err := &WordNotFoundError{Word: "xyzzy"}
	msg := err.Error()
	if !strings.Contains(msg, "xyzzy") {
		t.Errorf("error message should contain the word, got: %q", msg)
	}
	if strings.Contains(msg, "did you mean") {
		t.Errorf("error message should not mention suggestions when there are none, got: %q", msg)
	}
}

func TestWordNotFoundError_ImplementsError(t *testing.T) {
	// Compile-time check: *WordNotFoundError must satisfy the error interface.
	var _ error = (*WordNotFoundError)(nil)
}
