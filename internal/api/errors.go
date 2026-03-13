package api

import "fmt"

// WordNotFoundError is returned when MW responds with a suggestion list
// instead of entries, meaning the exact word was not found.
type WordNotFoundError struct {
	Word        string
	Suggestions []string
}

func (e *WordNotFoundError) Error() string {
	if len(e.Suggestions) == 0 {
		return fmt.Sprintf("word %q not found", e.Word)
	}
	return fmt.Sprintf("word %q not found; did you mean: %v", e.Word, e.Suggestions)
}
