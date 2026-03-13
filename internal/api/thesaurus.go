package api

// ThesEntry represents one entry from the MW Collegiate Thesaurus API.
type ThesEntry struct {
	Meta ThesMeta      `json:"meta"`
	Hwi  Hwi           `json:"hwi"`
	Fl   string        `json:"fl"`
	Def  []ThesDefBlock `json:"def"`
}

// ThesMeta holds thesaurus metadata including synonym and antonym lists.
// Syns and Ants are grouped by sense (each inner slice is one sense group).
type ThesMeta struct {
	ID        string     `json:"id"`
	Syns      [][]string `json:"syns"`
	Ants      [][]string `json:"ants"`
	Offensive bool       `json:"offensive"`
}

// ThesDefBlock is a thesaurus definition block.
type ThesDefBlock struct {
	Sseq [][][]any `json:"sseq"`
}

// Thesaurus fetches thesaurus entries for word.
// Both synonyms (Meta.Syns) and antonyms (Meta.Ants) are in the same response.
func (c *Client) Thesaurus(word string) ([]ThesEntry, error) {
	var entries []ThesEntry
	if err := c.get("thesaurus", word, c.apiKeyThesaurus, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
