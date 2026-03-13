package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/user/mweb/internal/api"
)

// ---- cleanMarkup ----

func TestCleanMarkup_BoldColon(t *testing.T) {
	got := cleanMarkup("{bc}a means of testing")
	if !strings.Contains(got, ":") {
		t.Errorf("expected {bc} to become ':', got %q", got)
	}
}

func TestCleanMarkup_Quotes(t *testing.T) {
	got := cleanMarkup("{ldquo}hello{rdquo}")
	if got != `"hello"` {
		t.Errorf("expected quoted string, got %q", got)
	}
}

func TestCleanMarkup_UnknownTokensRemoved(t *testing.T) {
	got := cleanMarkup("{it}word{/it}")
	if strings.Contains(got, "{") {
		t.Errorf("expected markup tokens removed, got %q", got)
	}
	if got != "word" {
		t.Errorf("expected %q, got %q", "word", got)
	}
}

func TestCleanMarkup_Trim(t *testing.T) {
	got := cleanMarkup("  hello  ")
	if got != "hello" {
		t.Errorf("expected trimmed string, got %q", got)
	}
}

// ---- helpers for building test fixtures ----

func makeDictEntry(hw, fl, sn, defText, date string) api.DictEntry {
	return api.DictEntry{
		Hwi:  api.Hwi{Hw: hw, Prs: []api.Prs{{Mw: "test-pron"}}},
		Fl:   fl,
		Date: date,
		Def: []api.DefBlock{
			{
				Sseq: [][][]any{
					{
						{
							"sense",
							map[string]any{
								"sn": sn,
								"dt": []any{
									[]any{"text", defText},
								},
							},
						},
					},
				},
			},
		},
	}
}

func makeThesEntry(hw, fl string, syns, ants [][]string) api.ThesEntry {
	return api.ThesEntry{
		Hwi:  api.Hwi{Hw: hw},
		Fl:   fl,
		Meta: api.ThesMeta{Syns: syns, Ants: ants},
	}
}

// ---- PrintDefinitions ----

func TestPrintDefinitions_PlainBasic(t *testing.T) {
	entries := []api.DictEntry{
		makeDictEntry("test", "noun", "1", "a means of testing", "14th century"),
	}
	var b strings.Builder
	if err := PrintDefinitions(&b, "test", entries, 5, FormatPlain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, "test (noun)") {
		t.Errorf("missing header, got:\n%s", out)
	}
	if !strings.Contains(out, "a means of testing") {
		t.Errorf("missing definition text, got:\n%s", out)
	}
	if strings.Contains(out, "14th century") {
		t.Errorf("date should not appear in plain output, got:\n%s", out)
	}
	if strings.Contains(out, "test-pron") {
		t.Errorf("pronunciation should not appear in plain output, got:\n%s", out)
	}
}

func TestPrintDefinitions_StripStressMarkers(t *testing.T) {
	entries := []api.DictEntry{
		makeDictEntry("ser*en*dip*i*ty", "noun", "1", "finding things not sought", "1754"),
	}
	var b strings.Builder
	PrintDefinitions(&b, "serendipity", entries, 5, FormatPlain)
	out := b.String()
	if strings.Contains(out, "*") {
		t.Errorf("headword should have stress markers stripped, got:\n%s", out)
	}
	if !strings.Contains(out, "serendipity") {
		t.Errorf("expected clean headword in output, got:\n%s", out)
	}
}

func TestPrintDefinitions_MaxDefs(t *testing.T) {
	sseq := [][][]any{}
	for i := 1; i <= 4; i++ {
		sseq = append(sseq, [][]any{
			{
				"sense",
				map[string]any{
					"sn": "",
					"dt": []any{
						[]any{"text", "definition text"},
					},
				},
			},
		})
	}
	entry := api.DictEntry{
		Hwi: api.Hwi{Hw: "run"},
		Fl:  "verb",
		Def: []api.DefBlock{{Sseq: sseq}},
	}

	var b strings.Builder
	PrintDefinitions(&b, "run", []api.DictEntry{entry}, 2, FormatPlain)
	out := b.String()
	count := strings.Count(out, "definition text")
	if count != 2 {
		t.Errorf("expected 2 definitions (maxDefs=2), got %d in:\n%s", count, out)
	}
}

func TestPrintDefinitions_NoEntries(t *testing.T) {
	var b strings.Builder
	PrintDefinitions(&b, "test", []api.DictEntry{}, 5, FormatPlain)
	if !strings.Contains(b.String(), "No definitions found") {
		t.Errorf("expected 'No definitions found', got: %q", b.String())
	}
}

func TestPrintDefinitions_SkipsPhraseEntries(t *testing.T) {
	entries := []api.DictEntry{
		makeDictEntry("run", "verb", "1", "to go faster than a walk", ""),
		makeDictEntry("run a fever", "verb phrase", "1", "to have a fever", ""),
		makeDictEntry("run-and-gun", "adjective", "1", "played at a fast pace", ""),
	}
	var b strings.Builder
	PrintDefinitions(&b, "run", entries, 5, FormatPlain)
	out := b.String()
	if strings.Contains(out, "run a fever") {
		t.Errorf("space-separated phrase should be filtered, got:\n%s", out)
	}
	if strings.Contains(out, "run-and-gun") {
		t.Errorf("hyphenated compound should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "run (verb)") {
		t.Errorf("exact match should still appear, got:\n%s", out)
	}
}

func TestPrintDefinitions_SkipsEntriesWithNoFl(t *testing.T) {
	entries := []api.DictEntry{
		{Hwi: api.Hwi{Hw: "test"}, Fl: "", Def: nil},
	}
	var b strings.Builder
	PrintDefinitions(&b, "test", entries, 5, FormatPlain)
	if !strings.Contains(b.String(), "No definitions found") {
		t.Errorf("expected 'No definitions found' for entry with empty Fl, got: %q", b.String())
	}
}

func TestPrintDefinitions_JSON(t *testing.T) {
	entries := []api.DictEntry{
		makeDictEntry("run", "verb", "1", "to go faster than a walk", ""),
	}
	var b strings.Builder
	if err := PrintDefinitions(&b, "run", entries, 5, FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded []api.DictEntry
	if err := json.Unmarshal([]byte(b.String()), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, b.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 JSON entry, got %d", len(decoded))
	}
}

func TestPrintDefinitions_SenseLabel(t *testing.T) {
	entries := []api.DictEntry{
		makeDictEntry("test", "noun", "1 a", "first sub-sense", ""),
	}
	var b strings.Builder
	PrintDefinitions(&b, "test", entries, 5, FormatPlain)
	if !strings.Contains(b.String(), "1 a.") {
		t.Errorf("expected sense label '1 a.' in output, got:\n%s", b.String())
	}
}

// ---- PrintSynonyms ----

func TestPrintSynonyms_Basic(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("hap*py", "adjective",
			[][]string{{"glad", "joyful", "blissful"}},
			nil,
		),
	}
	var b strings.Builder
	PrintSynonyms(&b, "happy", entries, FormatPlain)
	out := b.String()
	if !strings.Contains(out, `Synonyms for "happy"`) {
		t.Errorf("missing header, got:\n%s", out)
	}
	if !strings.Contains(out, "glad") {
		t.Errorf("missing synonym 'glad', got:\n%s", out)
	}
	if !strings.Contains(out, "joyful") {
		t.Errorf("missing synonym 'joyful', got:\n%s", out)
	}
}

func TestPrintSynonyms_MultipleSenseGroups(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("fast", "adjective",
			[][]string{{"quick", "speedy"}, {"fixed", "secure"}},
			nil,
		),
	}
	var b strings.Builder
	PrintSynonyms(&b, "fast", entries, FormatPlain)
	out := b.String()
	if !strings.Contains(out, "Sense 1:") {
		t.Errorf("expected 'Sense 1:' label for multi-sense, got:\n%s", out)
	}
	if !strings.Contains(out, "Sense 2:") {
		t.Errorf("expected 'Sense 2:' label for multi-sense, got:\n%s", out)
	}
}

func TestPrintSynonyms_FiltersMismatch(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("run", "verb", [][]string{{"sprint", "dash"}}, nil),
		makeThesEntry("run-down", "adjective", [][]string{{"tired", "weary"}}, nil),
	}
	var b strings.Builder
	PrintSynonyms(&b, "run", entries, FormatPlain)
	out := b.String()
	if strings.Contains(out, "run-down") {
		t.Errorf("compound entry should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "sprint") {
		t.Errorf("exact match should appear, got:\n%s", out)
	}
}

func TestPrintSynonyms_NoSynonyms(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("test", "noun", nil, nil),
	}
	var b strings.Builder
	PrintSynonyms(&b, "test", entries, FormatPlain)
	if !strings.Contains(b.String(), "No synonyms found") {
		t.Errorf("expected 'No synonyms found', got:\n%s", b.String())
	}
}

func TestPrintSynonyms_EmptyEntries(t *testing.T) {
	var b strings.Builder
	PrintSynonyms(&b, "test", []api.ThesEntry{}, FormatPlain)
	if !strings.Contains(b.String(), "No synonyms found") {
		t.Errorf("expected 'No synonyms found' for empty slice, got: %q", b.String())
	}
}

func TestPrintSynonyms_JSON(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("happy", "adjective", [][]string{{"glad"}}, nil),
	}
	var b strings.Builder
	if err := PrintSynonyms(&b, "happy", entries, FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded []api.ThesEntry
	if err := json.Unmarshal([]byte(b.String()), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, b.String())
	}
}

// ---- PrintAntonyms ----

func TestPrintAntonyms_Basic(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("hap*py", "adjective",
			nil,
			[][]string{{"sad", "unhappy", "miserable"}},
		),
	}
	var b strings.Builder
	PrintAntonyms(&b, "happy", entries, FormatPlain)
	out := b.String()
	if !strings.Contains(out, `Antonyms for "happy"`) {
		t.Errorf("missing header, got:\n%s", out)
	}
	if !strings.Contains(out, "sad") {
		t.Errorf("missing antonym 'sad', got:\n%s", out)
	}
}

func TestPrintAntonyms_FiltersMismatch(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("happy", "adjective", nil, [][]string{{"sad"}}),
		makeThesEntry("happy-go-lucky", "adjective", nil, [][]string{{"serious"}}),
	}
	var b strings.Builder
	PrintAntonyms(&b, "happy", entries, FormatPlain)
	out := b.String()
	if strings.Contains(out, "happy-go-lucky") {
		t.Errorf("compound entry should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "sad") {
		t.Errorf("exact match should appear, got:\n%s", out)
	}
}

func TestPrintAntonyms_NoAntonyms(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("test", "noun", nil, nil),
	}
	var b strings.Builder
	PrintAntonyms(&b, "test", entries, FormatPlain)
	if !strings.Contains(b.String(), "No antonyms found") {
		t.Errorf("expected 'No antonyms found', got: %q", b.String())
	}
}

func TestPrintAntonyms_EmptyEntries(t *testing.T) {
	var b strings.Builder
	PrintAntonyms(&b, "test", []api.ThesEntry{}, FormatPlain)
	if !strings.Contains(b.String(), "No antonyms found") {
		t.Errorf("expected 'No antonyms found' for empty slice, got: %q", b.String())
	}
}

func TestPrintAntonyms_JSON(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("happy", "adjective", nil, [][]string{{"sad"}}),
	}
	var b strings.Builder
	if err := PrintAntonyms(&b, "happy", entries, FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded []api.ThesEntry
	if err := json.Unmarshal([]byte(b.String()), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, b.String())
	}
}

// ---- extractSenses ----

func TestExtractSenses_BasicSense(t *testing.T) {
	defs := []api.DefBlock{
		{
			Sseq: [][][]any{
				{
					{
						"sense",
						map[string]any{
							"sn": "1",
							"dt": []any{
								[]any{"text", "the act of testing"},
							},
						},
					},
				},
			},
		},
	}
	senses := extractSenses(defs)
	if len(senses) != 1 {
		t.Fatalf("expected 1 sense, got %d", len(senses))
	}
	if senses[0].label != "1" {
		t.Errorf("label = %q, want \"1\"", senses[0].label)
	}
	if senses[0].text != "the act of testing" {
		t.Errorf("text = %q, want \"the act of testing\"", senses[0].text)
	}
}

func TestExtractSenses_BoldColonInText(t *testing.T) {
	defs := []api.DefBlock{
		{
			Sseq: [][][]any{
				{
					{
						"sense",
						map[string]any{
							"sn": "1",
							"dt": []any{
								[]any{"text", "{bc}an examination"},
							},
						},
					},
				},
			},
		},
	}
	senses := extractSenses(defs)
	if len(senses) != 1 {
		t.Fatalf("expected 1 sense, got %d", len(senses))
	}
	if !strings.Contains(senses[0].text, ":") {
		t.Errorf("expected {bc} to become ':', got %q", senses[0].text)
	}
	if !strings.Contains(senses[0].text, "an examination") {
		t.Errorf("expected definition text, got %q", senses[0].text)
	}
}

func TestExtractSenses_BSTag(t *testing.T) {
	defs := []api.DefBlock{
		{
			Sseq: [][][]any{
				{
					{
						"bs",
						map[string]any{
							"sense": map[string]any{
								"sn": "2",
								"dt": []any{
									[]any{"text", "binding substitute sense"},
								},
							},
						},
					},
				},
			},
		},
	}
	senses := extractSenses(defs)
	if len(senses) != 1 {
		t.Fatalf("expected 1 sense from 'bs' tag, got %d", len(senses))
	}
	if senses[0].text != "binding substitute sense" {
		t.Errorf("text = %q, want \"binding substitute sense\"", senses[0].text)
	}
}

func TestExtractSenses_UnknownTagSkipped(t *testing.T) {
	defs := []api.DefBlock{
		{
			Sseq: [][][]any{
				{
					{"pseq", map[string]any{"something": "value"}},
				},
			},
		},
	}
	senses := extractSenses(defs)
	if len(senses) != 0 {
		t.Errorf("expected 0 senses for unknown tag, got %d", len(senses))
	}
}

func TestExtractSenses_EmptyDt(t *testing.T) {
	defs := []api.DefBlock{
		{
			Sseq: [][][]any{
				{
					{
						"sense",
						map[string]any{
							"sn": "1",
							"dt": []any{},
						},
					},
				},
			},
		},
	}
	senses := extractSenses(defs)
	if len(senses) != 0 {
		t.Errorf("expected 0 senses for empty dt, got %d", len(senses))
	}
}
