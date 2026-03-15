package output

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/user/mweb/internal/api"
)

// Format controls output rendering.
type Format string

const (
	FormatPlain Format = "plain"
	FormatJSON  Format = "json"
)

// mwMarkup strips MW inline markup tokens like {bc}, {it}, {b}, {ldquo}, etc.
var mwMarkup = regexp.MustCompile(`\{[^}]+\}`)

// extractTokenText returns the display text embedded in a MW cross-reference
// token of the form {tag|text|...}. Returns "" for tokens without a pipe.
func extractTokenText(token string) string {
	// strip surrounding braces
	inner := token[1 : len(token)-1]
	idx := strings.Index(inner, "|")
	if idx < 0 {
		return ""
	}
	// text is between first and second pipe (or end)
	rest := inner[idx+1:]
	end := strings.Index(rest, "|")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func cleanMarkup(s string) string {
	s = mwMarkup.ReplaceAllStringFunc(s, func(m string) string {
		switch {
		case m == "{bc}":
			return ": "
		case m == "{ldquo}":
			return "\""
		case m == "{rdquo}":
			return "\""
		// Cross-reference tokens: {sx|word||}, {a_link|word}, {d_link|word|id}, etc.
		// The display text is always the first segment after the opening pipe.
		case strings.Contains(m, "|"):
			return extractTokenText(m)
		default:
			return ""
		}
	})
	// Collapse any double-spaces left by stripped tokens and trim.
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// headword returns the cleaned headword for an entry (stress markers removed).
func headword(hw string) string {
	return strings.ReplaceAll(hw, "*", "")
}

// matchesWord reports whether an entry is relevant to the searched word.
// It first checks the headword directly (to filter unrelated phrases/compounds),
// then falls back to the stems list to handle inflected forms — e.g. searching
// "nucleating" matches the "nucleate" entry because "nucleating" is in its stems.
func matchesWord(hw string, stems []string, word string) bool {
	if strings.EqualFold(headword(hw), word) {
		return true
	}
	for _, stem := range stems {
		if strings.EqualFold(stem, word) {
			return true
		}
	}
	return false
}

// PrintDefinitions renders dictionary entries for word to w.
func PrintDefinitions(w io.Writer, word string, entries []api.DictEntry, maxDefs int, format Format) error {
	if format == FormatJSON {
		return printJSON(w, entries)
	}

	// Collect matching entries and detect whether we fell back to stem matching.
	type match struct {
		entry      api.DictEntry
		stemMatch  bool
	}
	var matches []match
	for _, e := range entries {
		if e.Fl == "" || !matchesWord(e.Hwi.Hw, e.Meta.Stems, word) {
			continue
		}
		if len(extractSenses(e.Def)) == 0 {
			continue
		}
		isStem := !strings.EqualFold(headword(e.Hwi.Hw), word)
		matches = append(matches, match{e, isStem})
	}

	if len(matches) == 0 {
		fmt.Fprintln(w, "No definitions found.")
		return nil
	}

	// If every result came from stem matching, show what we resolved to.
	if matches[0].stemMatch {
		fmt.Fprintf(w, "(Showing results for %q)\n", headword(matches[0].entry.Hwi.Hw))
	}

	for _, m := range matches {
		e := m.entry
		fmt.Fprintf(w, "\n%s (%s)\n", headword(e.Hwi.Hw), e.Fl)
		printed := 0
		for _, s := range extractSenses(e.Def) {
			if maxDefs > 0 && printed >= maxDefs {
				break
			}
			printed++
			label := s.label
			if label == "" {
				label = fmt.Sprintf("%d", printed)
			}
			fmt.Fprintf(w, "  %s. %s\n", label, s.text)
		}
	}
	return nil
}

// PrintSynonyms renders synonym lists for word to w.
func PrintSynonyms(w io.Writer, word string, entries []api.ThesEntry, format Format) error {
	if format == FormatJSON {
		return printJSON(w, entries)
	}

	found := false
	stemFallbackNote := false
	for _, e := range entries {
		if len(e.Meta.Syns) == 0 || !matchesWord(e.Hwi.Hw, e.Meta.Stems, word) {
			continue
		}
		if !found && !strings.EqualFold(headword(e.Hwi.Hw), word) {
			stemFallbackNote = true
		}
		found = true
		hw := headword(e.Hwi.Hw)
		if stemFallbackNote {
			fmt.Fprintf(w, "(Showing results for %q)\n", hw)
			stemFallbackNote = false
		}
		fl := ""
		if e.Fl != "" {
			fl = fmt.Sprintf(" (%s)", e.Fl)
		}
		fmt.Fprintf(w, "\nSynonyms for \"%s\"%s:\n", hw, fl)
		for i, group := range e.Meta.Syns {
			if len(group) == 0 {
				continue
			}
			if len(e.Meta.Syns) > 1 {
				fmt.Fprintf(w, "  Sense %d: %s\n", i+1, strings.Join(group, ", "))
			} else {
				fmt.Fprintf(w, "  %s\n", strings.Join(group, ", "))
			}
		}
	}

	if !found {
		fmt.Fprintf(w, "No synonyms found for \"%s\".\n", word)
	}
	return nil
}

// PrintAntonyms renders antonym lists for word to w.
func PrintAntonyms(w io.Writer, word string, entries []api.ThesEntry, format Format) error {
	if format == FormatJSON {
		return printJSON(w, entries)
	}

	found := false
	stemFallbackNote := false
	for _, e := range entries {
		if len(e.Meta.Ants) == 0 || !matchesWord(e.Hwi.Hw, e.Meta.Stems, word) {
			continue
		}
		if !found && !strings.EqualFold(headword(e.Hwi.Hw), word) {
			stemFallbackNote = true
		}
		found = true
		hw := headword(e.Hwi.Hw)
		if stemFallbackNote {
			fmt.Fprintf(w, "(Showing results for %q)\n", hw)
			stemFallbackNote = false
		}
		fl := ""
		if e.Fl != "" {
			fl = fmt.Sprintf(" (%s)", e.Fl)
		}
		fmt.Fprintf(w, "\nAntonyms for \"%s\"%s:\n", hw, fl)
		for i, group := range e.Meta.Ants {
			if len(group) == 0 {
				continue
			}
			if len(e.Meta.Ants) > 1 {
				fmt.Fprintf(w, "  Sense %d: %s\n", i+1, strings.Join(group, ", "))
			} else {
				fmt.Fprintf(w, "  %s\n", strings.Join(group, ", "))
			}
		}
	}

	if !found {
		fmt.Fprintf(w, "No antonyms found for \"%s\".\n", word)
	}
	return nil
}

// printJSON marshals v as indented JSON to w.
func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// sense holds a parsed definition sense label and text.
type sense struct {
	label string
	text  string
}

// extractSenses walks the sseq structure of MW definition blocks and returns
// flat sense label + text pairs. MW's sseq is [][][]any where each
// leaf is ["sense", {...}], ["bs", {...}], ["pseq", [...]], or ["sen", {...}].
func extractSenses(defs []api.DefBlock) []sense {
	var senses []sense
	for _, block := range defs {
		for _, sseqEntry := range block.Sseq {
			for _, item := range sseqEntry {
				senses = append(senses, extractSenseItem(item)...)
			}
		}
	}
	return senses
}

// extractSenseItem processes a single [tag, data] item from an sseq entry.
// Handles "sense", "bs" (binding substitute), and "pseq" (parenthesized sequence).
// "sen" (truncated sense with no definition) and unknown tags are skipped.
func extractSenseItem(item []any) []sense {
	if len(item) < 2 {
		return nil
	}
	tag, ok := item[0].(string)
	if !ok {
		return nil
	}
	switch tag {
	case "sense", "bs":
		senseObj, ok := item[1].(map[string]any)
		if !ok {
			return nil
		}
		if tag == "bs" {
			inner, ok := senseObj["sense"].(map[string]any)
			if !ok {
				return nil
			}
			senseObj = inner
		}
		return parseSenseObj(senseObj)
	case "pseq":
		// Parenthesized sequence: data is []any of [tag, data] pairs.
		nested, ok := item[1].([]any)
		if !ok {
			return nil
		}
		var out []sense
		for _, ni := range nested {
			nItem, ok := ni.([]any)
			if !ok {
				continue
			}
			out = append(out, extractSenseItem(nItem)...)
		}
		return out
	}
	return nil
}

// parseSenseObj extracts senses from a decoded MW sense object. If a divided
// sense (sdsense) is present — e.g. "especially X" — it is appended inline
// to the main definition separated by a semicolon.
func parseSenseObj(senseObj map[string]any) []sense {
	label := ""
	if sn, ok := senseObj["sn"].(string); ok {
		label = sn
	}

	text := extractDt(senseObj)

	// Append sdsense (divided sense: "especially", "also", "broadly", etc.)
	if sds, ok := senseObj["sdsense"].(map[string]any); ok {
		sdText := extractDt(sds)
		if sdText != "" {
			sd := "also"
			if sdStr, ok := sds["sd"].(string); ok {
				sd = sdStr
			}
			if text != "" {
				text = text + "; " + sd + " " + sdText
			} else {
				text = sd + " " + sdText
			}
		}
	}

	if text == "" {
		return nil
	}
	return []sense{{label: label, text: text}}
}

// PrintExamples renders verbal illustration (example usage) sentences for word to w.
// Examples are sourced from the collegiate dictionary API, not the thesaurus.
func PrintExamples(w io.Writer, word string, entries []api.DictEntry, format Format) error {
	if format == FormatJSON {
		return printJSON(w, entries)
	}

	type exEntry struct {
		hw       string
		fl       string
		stemMatch bool
		examples []string
	}
	var matches []exEntry
	for _, e := range entries {
		if e.Fl == "" || !matchesWord(e.Hwi.Hw, e.Meta.Stems, word) {
			continue
		}
		exs := extractExamplesFromDefs(e.Def)
		if len(exs) == 0 {
			continue
		}
		isStem := !strings.EqualFold(headword(e.Hwi.Hw), word)
		matches = append(matches, exEntry{headword(e.Hwi.Hw), e.Fl, isStem, exs})
	}

	if len(matches) == 0 {
		fmt.Fprintf(w, "No example usage found for \"%s\".\n", word)
		return nil
	}

	if matches[0].stemMatch {
		fmt.Fprintf(w, "(Showing results for %q)\n", matches[0].hw)
	}

	for _, m := range matches {
		fmt.Fprintf(w, "\n%s (%s):\n", m.hw, m.fl)
		for i, ex := range m.examples {
			fmt.Fprintf(w, "  %d. %s\n", i+1, ex)
		}
	}
	return nil
}

// extractExamplesFromDefs walks the sseq structure of MW definition blocks and
// returns all verbal illustration ("vis") texts across all senses.
func extractExamplesFromDefs(defs []api.DefBlock) []string {
	var out []string
	for _, block := range defs {
		for _, sseqEntry := range block.Sseq {
			for _, item := range sseqEntry {
				out = append(out, extractExamplesFromItem(item)...)
			}
		}
	}
	return out
}

// extractExamplesFromItem mirrors extractSenseItem but collects vis texts.
func extractExamplesFromItem(item []any) []string {
	if len(item) < 2 {
		return nil
	}
	tag, ok := item[0].(string)
	if !ok {
		return nil
	}
	switch tag {
	case "sense", "bs":
		senseObj, ok := item[1].(map[string]any)
		if !ok {
			return nil
		}
		if tag == "bs" {
			inner, ok := senseObj["sense"].(map[string]any)
			if !ok {
				return nil
			}
			senseObj = inner
		}
		return extractVis(senseObj)
	case "pseq":
		nested, ok := item[1].([]any)
		if !ok {
			return nil
		}
		var out []string
		for _, ni := range nested {
			nItem, ok := ni.([]any)
			if !ok {
				continue
			}
			out = append(out, extractExamplesFromItem(nItem)...)
		}
		return out
	}
	return nil
}

// extractVis pulls verbal illustration texts from a sense object's "dt" field.
// Each vis item has the shape {"t": "example sentence", "aq": {...}}.
func extractVis(senseObj map[string]any) []string {
	dtRaw, ok := senseObj["dt"]
	if !ok {
		return nil
	}
	dt, ok := dtRaw.([]any)
	if !ok {
		return nil
	}

	var out []string
	for _, item := range dt {
		pair, ok := item.([]any)
		if !ok || len(pair) < 2 || pair[0] != "vis" {
			continue
		}
		visList, ok := pair[1].([]any)
		if !ok {
			continue
		}
		for _, v := range visList {
			vObj, ok := v.(map[string]any)
			if !ok {
				continue
			}
			t, ok := vObj["t"].(string)
			if !ok || t == "" {
				continue
			}
			out = append(out, cleanMarkup(t))
		}
	}
	return out
}

// extractDt pulls the defining text from a sense object's "dt" field.
func extractDt(senseObj map[string]any) string {
	dtRaw, ok := senseObj["dt"]
	if !ok {
		return ""
	}
	dt, ok := dtRaw.([]any)
	if !ok {
		return ""
	}

	var parts []string
	for _, item := range dt {
		pair, ok := item.([]any)
		if !ok || len(pair) < 2 {
			continue
		}
		typ, ok := pair[0].(string)
		if !ok || typ != "text" {
			continue
		}
		text, ok := pair[1].(string)
		if !ok {
			continue
		}
		parts = append(parts, cleanMarkup(text))
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}
