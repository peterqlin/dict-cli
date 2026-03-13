package api

// DictEntry represents one entry from the MW Collegiate Dictionary API.
// Only the fields relevant to `def` output are mapped.
type DictEntry struct {
	Meta DictMeta   `json:"meta"`
	Hwi  Hwi        `json:"hwi"`
	Fl   string     `json:"fl"`   // functional label: "noun", "verb", etc.
	Def  []DefBlock `json:"def"`
	Date string     `json:"date"` // first known use
}

// DictMeta holds entry metadata.
type DictMeta struct {
	ID        string   `json:"id"`
	Stems     []string `json:"stems"`
	Offensive bool     `json:"offensive"`
}

// Hwi is headword info.
type Hwi struct {
	Hw  string `json:"hw"`  // headword with stress markers (e.g. "se*ren*dip*i*ty")
	Prs []Prs  `json:"prs"` // pronunciations
}

// Prs holds a single pronunciation.
type Prs struct {
	Mw string `json:"mw"` // written pronunciation
}

// DefBlock is a definition block, which contains a sense sequence.
type DefBlock struct {
	Vd   string              `json:"vd"`   // verb divider ("transitive verb")
	Sseq [][][]any `json:"sseq"` // nested sense sequences (MW's complex structure)
}

// Define fetches dictionary entries for word from the Collegiate Dictionary.
// Multiple entries are returned for homographs.
func (c *Client) Define(word string) ([]DictEntry, error) {
	var entries []DictEntry
	if err := c.get("collegiate", word, c.apiKeyDict, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
