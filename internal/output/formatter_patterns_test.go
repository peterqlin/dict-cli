package output

// Real-world MW API structural pattern tests.
//
// These tests mirror the JSON shapes that the MW Collegiate Dictionary API
// actually returns for specific words, catching regressions in sseq parsing.
// Words covered: dynamo (pseq), paltry (sx cross-ref), run (homographs +
// phrase filter), bark (homographs), nucleating (stem fallback), cranky
// (sdsense), set/bear/bank (complex sseq), along with markup-stripping,
// vis/uns skipping, bs tag, sen tag, and thesaurus sense-group patterns.

import (
	"strings"
	"testing"

	"github.com/user/mweb/internal/api"
)

// ── sseq construction helpers ────────────────────────────────────────────────

// sseqOf wraps items into a single sseqEntry.
func sseqOf(items ...[]any) [][][]any {
	entry := make([][]any, len(items))
	for i, item := range items {
		entry[i] = item
	}
	return [][][]any{entry}
}

// senseItem builds a ["sense", {sn, dt}] sseq item.
func senseItem(sn, defText string) []any {
	return []any{"sense", map[string]any{
		"sn": sn,
		"dt": []any{[]any{"text", defText}},
	}}
}

// bsItem builds a ["bs", {"sense": {sn, dt}}] sseq item (binding substitute).
func bsItem(sn, defText string) []any {
	return []any{"bs", map[string]any{
		"sense": map[string]any{
			"sn": sn,
			"dt": []any{[]any{"text", defText}},
		},
	}}
}

// senItem builds a ["sen", {sn}] sseq item (truncated sense — no definition).
func senItem(sn string) []any {
	return []any{"sen", map[string]any{"sn": sn}}
}

// pseqItem builds a ["pseq", [...items]] sseq item.
func pseqItem(nestedItems ...[]any) []any {
	nested := make([]any, len(nestedItems))
	for i, ni := range nestedItems {
		nested[i] = ni
	}
	return []any{"pseq", nested}
}

// senseWithSdsense builds a ["sense", {sn, dt, sdsense}] sseq item.
func senseWithSdsense(sn, defText, sd, sdText string) []any {
	return []any{"sense", map[string]any{
		"sn": sn,
		"dt": []any{[]any{"text", defText}},
		"sdsense": map[string]any{
			"sd": sd,
			"dt": []any{[]any{"text", sdText}},
		},
	}}
}

// senseWithVisDt builds a ["sense", {sn, dt}] where dt has both text and vis.
func senseWithVisDt(sn, defText, visText string) []any {
	return []any{"sense", map[string]any{
		"sn": sn,
		"dt": []any{
			[]any{"text", defText},
			[]any{"vis", []any{map[string]any{"t": visText}}},
		},
	}}
}

// senseWithUnsDt builds a ["sense", {sn, dt}] where dt has text and uns.
func senseWithUnsDt(sn, defText string) []any {
	return []any{"sense", map[string]any{
		"sn": sn,
		"dt": []any{
			[]any{"text", defText},
			[]any{"uns", []any{[]any{[]any{"text", "usage note text"}}}},
		},
	}}
}

// entryWith builds a DictEntry with the given sseq.
func entryWith(hw, fl string, stems []string, sseq [][][]any) api.DictEntry {
	e := api.DictEntry{
		Hwi: api.Hwi{Hw: hw},
		Fl:  fl,
		Def: []api.DefBlock{{Sseq: sseq}},
	}
	e.Meta.Stems = stems
	return e
}

// entryMultiBlock builds a DictEntry with two def blocks (e.g. transitive +
// intransitive verb senses).
func entryMultiBlock(hw, fl string, blocks ...api.DefBlock) api.DictEntry {
	return api.DictEntry{
		Hwi: api.Hwi{Hw: hw},
		Fl:  fl,
		Def: blocks,
	}
}

// ── pseq (parenthesized sequence) ───────────────────────────────────────────

// TestExtractSenses_Pseq_Dynamo mirrors the real MW response for "dynamo",
// which groups senses 1a and 1b inside a pseq block. Without pseq support,
// PrintDefinitions returns "No definitions found."
func TestExtractSenses_Pseq_Dynamo(t *testing.T) {
	entry := entryWith("dy*na*mo", "noun", []string{"dynamo", "dynamos"},
		sseqOf(
			pseqItem(
				senseItem("1 a", "{bc}a power generator"),
				senseItem("b", "{bc}a forceful energetic person"),
			),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "dynamo", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if strings.Contains(out, "No definitions found") {
		t.Fatalf("dynamo returned no definitions; pseq senses were dropped:\n%s", out)
	}
	if !strings.Contains(out, "power generator") {
		t.Errorf("missing sense 1a (power generator):\n%s", out)
	}
	if !strings.Contains(out, "forceful energetic person") {
		t.Errorf("missing sense 1b (forceful energetic person):\n%s", out)
	}
}

// TestExtractSenses_Pseq_NestedMixed verifies a mixed sseq where some
// top-level items are plain senses and others are pseq blocks.
func TestExtractSenses_Pseq_NestedMixed(t *testing.T) {
	entry := entryWith("spring", "noun", []string{"spring", "springs"},
		sseqOf(
			senseItem("1", "{bc}a source of water"),
			pseqItem(
				senseItem("2 a", "{bc}the season after winter"),
				senseItem("b", "{bc}a time of growth"),
			),
			senseItem("3", "{bc}a mechanical coil"),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "spring", []api.DictEntry{entry}, 10, FormatPlain)
	out := b.String()

	for _, want := range []string{"source of water", "season after winter", "time of growth", "mechanical coil"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

// ── sdsense (divided sense) ──────────────────────────────────────────────────

// TestExtractSenses_Sdsense mirrors words like "cranky" or "specifically"
// where a main sense is refined by "especially X" or "broadly Y".
func TestExtractSenses_Sdsense_Especially(t *testing.T) {
	entry := entryWith("cran*ky", "adjective", []string{"cranky", "crankier", "crankiest"},
		sseqOf(
			senseWithSdsense("1",
				"{bc}easily irritated",
				"especially",
				"{bc}habitually ill-tempered",
			),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "cranky", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "irritated") {
		t.Errorf("main sdsense definition missing:\n%s", out)
	}
	if !strings.Contains(out, "especially") {
		t.Errorf("sdsense divider word missing:\n%s", out)
	}
	if !strings.Contains(out, "ill-tempered") {
		t.Errorf("sdsense refinement text missing:\n%s", out)
	}
}

func TestExtractSenses_Sdsense_Also(t *testing.T) {
	entry := entryWith("bank", "noun", nil,
		sseqOf(
			senseWithSdsense("1",
				"{bc}a financial institution",
				"also",
				"{bc}its building",
			),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "bank", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "financial institution") {
		t.Errorf("missing main definition:\n%s", out)
	}
	if !strings.Contains(out, "also") {
		t.Errorf("missing 'also' divider:\n%s", out)
	}
	if !strings.Contains(out, "building") {
		t.Errorf("missing sdsense text:\n%s", out)
	}
}

func TestExtractSenses_Sdsense_OnlySubdef(t *testing.T) {
	// A sense whose main dt is empty but sdsense has content should still appear.
	entry := entryWith("word", "noun", nil,
		sseqOf(
			[]any{"sense", map[string]any{
				"sn": "1",
				"dt": []any{},
				"sdsense": map[string]any{
					"sd": "especially",
					"dt": []any{[]any{"text", "{bc}a specific meaning"}},
				},
			}},
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "word", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "specific meaning") {
		t.Errorf("sdsense-only sense should appear:\n%s", out)
	}
}

// ── cross-reference token rendering (regression: paltry) ────────────────────

// TestPrintDef_Paltry_CrossRef is a regression test for the bug where
// {sx|word||} tokens were stripped entirely, leaving empty definition lines.
func TestPrintDef_Paltry_CrossRef(t *testing.T) {
	entry := entryWith("pal*try", "adjective", []string{"paltry", "paltrier", "paltriest"},
		sseqOf(
			senseItem("1", "{bc}{sx|inferior||}, {sx|trashy||}"),
			senseItem("2", "{bc}{sx|mean||}, {sx|despicable||}"),
			senseItem("3", "{bc}trivially small : {sx|meager||}"),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "paltry", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if strings.Contains(out, "No definitions found") {
		t.Fatalf("paltry cross-ref senses were all empty:\n%s", out)
	}
	for _, want := range []string{"inferior", "trashy", "mean", "despicable", "meager"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing cross-ref word %q:\n%s", want, out)
		}
	}
	// No raw markup tokens should remain.
	if strings.Contains(out, "{sx|") || strings.Contains(out, "{bc}") {
		t.Errorf("raw markup tokens in output:\n%s", out)
	}
}

// TestPrintDef_NoEmptyLines asserts that no definition line is blank or
// contains only punctuation/whitespace (the original paltry symptom).
func TestPrintDef_NoEmptyLines(t *testing.T) {
	entry := entryWith("pal*try", "adjective", nil,
		sseqOf(
			senseItem("1", "{bc}{sx|inferior||}"),
			senseItem("2", "{bc}{sx|mean||}"),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "paltry", []api.DictEntry{entry}, 5, FormatPlain)

	for _, line := range strings.Split(b.String(), "\n") {
		trimmed := strings.TrimSpace(line)
		// Definition lines start with spaces then "N."; catch lines that are
		// just punctuation like ": ," or ":".
		if strings.HasPrefix(trimmed, "1.") || strings.HasPrefix(trimmed, "2.") {
			content := strings.TrimLeft(trimmed, "0123456789. ")
			content = strings.TrimLeft(content, ":, ")
			if content == "" {
				t.Errorf("definition line is effectively empty: %q\nfull output:\n%s", line, b.String())
			}
		}
	}
}

// ── markup token stripping ───────────────────────────────────────────────────

func TestCleanMarkup_AllTokenTypes(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		// Already-tested tokens (regression).
		{"{bc}definition", ": definition"},
		{"{ldquo}hello{rdquo}", `"hello"`},
		// Italic / bold formatting tokens — stripped.
		{"{it}word{/it}", "word"},
		{"{b}word{/b}", "word"},
		{"{sc}word{/sc}", "word"},
		// Subscript/superscript: the wrapping tokens strip but the content stays.
		// e.g. "CO{inf}2{/inf}" → "CO2" (still readable without formatting).
		{"{inf}2{/inf}", "2"},
		{"{sup}2{/sup}", "2"},
		// Cross-reference link tokens — display text extracted.
		{"{sx|inferior||}", "inferior"},
		{"{sx|run|run:1|}", "run"},
		{"{a_link|serendipity}", "serendipity"},
		{"{d_link|word|word:1}", "word"},
		{"{i_link|botany}", "botany"},
		{"{et_link|aqua|aqua}", "aqua"},
		{"{mat|happiness|happiness:1}", "happiness"},
		// Gloss and phrase tokens.
		{"{gloss|a group of islands}", "a group of islands"},
		{"{phrase|in a word}", "in a word"},
		// Wi (word in definition).
		{"{wi}run{/wi}", "run"},
		// Multiple tokens in one string.
		{"{bc}{sx|inferior||}, {sx|trashy||}", ": inferior, trashy"},
		// Whitespace collapse.
		{"{it}{/it}  hello  {b}{/b}", "hello"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := cleanMarkup(tc.input)
			if got != tc.want {
				t.Errorf("cleanMarkup(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ── bs (binding substitute) and sen (truncated sense) tags ──────────────────

func TestExtractSenses_BsTag(t *testing.T) {
	entry := entryWith("test", "noun", nil,
		sseqOf(bsItem("1", "{bc}a binding substitute sense")),
	)

	var b strings.Builder
	PrintDefinitions(&b, "test", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "binding substitute sense") {
		t.Errorf("bs tag sense missing:\n%s", out)
	}
}

func TestExtractSenses_SenTagSkipped(t *testing.T) {
	// A sen item has a label but no definition text; it should not produce output.
	entry := entryWith("test", "noun", nil,
		sseqOf(
			senItem("1"),
			senseItem("2", "{bc}a real definition"),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "test", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "real definition") {
		t.Errorf("sense after sen tag missing:\n%s", out)
	}
}

// ── vis and uns in dt are skipped ───────────────────────────────────────────

func TestExtractSenses_VisSkipped(t *testing.T) {
	entry := entryWith("run", "verb", nil,
		sseqOf(senseWithVisDt("1", "{bc}to move quickly", "she ran to the store")),
	)

	var b strings.Builder
	PrintDefinitions(&b, "run", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "move quickly") {
		t.Errorf("definition text missing:\n%s", out)
	}
	// Verbal illustration text should not appear in plain output.
	if strings.Contains(out, "she ran to the store") {
		t.Errorf("vis text should be excluded from output:\n%s", out)
	}
}

func TestExtractSenses_UnsSkipped(t *testing.T) {
	entry := entryWith("shall", "verb", nil,
		sseqOf(senseWithUnsDt("1", "{bc}used to express the future")),
	)

	var b strings.Builder
	PrintDefinitions(&b, "shall", []api.DictEntry{entry}, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "future") {
		t.Errorf("definition text missing:\n%s", out)
	}
	if strings.Contains(out, "usage note text") {
		t.Errorf("uns text should not appear in output:\n%s", out)
	}
}

// ── homographs ───────────────────────────────────────────────────────────────

func TestPrintDef_Homographs_Bark(t *testing.T) {
	entries := []api.DictEntry{
		entryWith("bark", "noun", []string{"bark", "barks"},
			sseqOf(senseItem("1", "{bc}the tough outer covering of a woody plant"))),
		entryWith("bark", "verb", []string{"bark", "barked", "barking"},
			sseqOf(senseItem("1", "{bc}to make the short loud cry of a dog"))),
	}

	var b strings.Builder
	PrintDefinitions(&b, "bark", entries, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "tough outer covering") {
		t.Errorf("bark (noun) missing:\n%s", out)
	}
	if !strings.Contains(out, "short loud cry") {
		t.Errorf("bark (verb) missing:\n%s", out)
	}
}

func TestPrintDef_Homographs_Lie(t *testing.T) {
	entries := []api.DictEntry{
		entryWith("lie", "verb", []string{"lie", "lay", "lain", "lying"},
			sseqOf(senseItem("1", "{bc}to be in a horizontal position"))),
		entryWith("lie", "verb", []string{"lie", "lied", "lying"},
			sseqOf(senseItem("1", "{bc}to make an untrue statement"))),
	}

	var b strings.Builder
	PrintDefinitions(&b, "lie", entries, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "horizontal position") {
		t.Errorf("lie (recline) missing:\n%s", out)
	}
	if !strings.Contains(out, "untrue statement") {
		t.Errorf("lie (deceive) missing:\n%s", out)
	}
}

func TestPrintDef_Homographs_Bear(t *testing.T) {
	entries := []api.DictEntry{
		entryWith("bear", "noun", []string{"bear", "bears"},
			sseqOf(senseItem("1", "{bc}a large heavy mammal"))),
		entryWith("bear", "verb", []string{"bear", "bore", "borne", "bearing"},
			sseqOf(senseItem("1", "{bc}to carry or support"))),
	}

	var b strings.Builder
	PrintDefinitions(&b, "bear", entries, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "large heavy mammal") {
		t.Errorf("bear (noun) missing:\n%s", out)
	}
	if !strings.Contains(out, "carry or support") {
		t.Errorf("bear (verb) missing:\n%s", out)
	}
}

// ── phrase filtering ─────────────────────────────────────────────────────────

func TestPrintDef_PhraseFilter_Run(t *testing.T) {
	entries := []api.DictEntry{
		entryWith("run", "verb", nil,
			sseqOf(senseItem("1", "{bc}to go faster than a walk"))),
		entryWith("run", "noun", nil,
			sseqOf(senseItem("1", "{bc}an act of running"))),
		entryWith("run a fever", "verb phrase", nil,
			sseqOf(senseItem("1", "{bc}to have a fever"))),
		entryWith("run-and-gun", "adjective", nil,
			sseqOf(senseItem("1", "{bc}played at a fast pace"))),
		entryWith("run-down", "adjective", nil,
			sseqOf(senseItem("1", "{bc}being in poor condition"))),
	}

	var b strings.Builder
	PrintDefinitions(&b, "run", entries, 10, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "go faster than a walk") {
		t.Errorf("run (verb) definition missing:\n%s", out)
	}
	if !strings.Contains(out, "act of running") {
		t.Errorf("run (noun) definition missing:\n%s", out)
	}
	for _, banned := range []string{"run a fever", "run-and-gun", "run-down", "have a fever", "fast pace", "poor condition"} {
		if strings.Contains(out, banned) {
			t.Errorf("phrase/compound %q should be filtered:\n%s", banned, out)
		}
	}
}

func TestPrintSyn_PhraseFilter(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("run", "verb", [][]string{{"sprint", "dash", "race"}}, nil),
		makeThesEntry("run-down", "adjective", [][]string{{"tired", "weary"}}, nil),
		makeThesEntry("run amok", "verb phrase", [][]string{{"rampage"}}, nil),
	}

	var b strings.Builder
	PrintSynonyms(&b, "run", entries, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "sprint") {
		t.Errorf("run synonyms missing:\n%s", out)
	}
	for _, banned := range []string{"run-down", "run amok", "tired", "rampage"} {
		if strings.Contains(out, banned) {
			t.Errorf("%q should be filtered from synonyms:\n%s", banned, out)
		}
	}
}

// ── stem fallback ────────────────────────────────────────────────────────────

func TestPrintDef_StemFallback_Running(t *testing.T) {
	entries := []api.DictEntry{
		entryWith("run", "verb", []string{"run", "ran", "runs", "running"},
			sseqOf(senseItem("1", "{bc}to go faster than a walk"))),
	}

	var b strings.Builder
	PrintDefinitions(&b, "running", entries, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, `Showing results for "run"`) {
		t.Errorf("stem fallback note missing:\n%s", out)
	}
	if !strings.Contains(out, "go faster than a walk") {
		t.Errorf("definition missing after stem fallback:\n%s", out)
	}
}

func TestPrintDef_StemFallback_BestToGood(t *testing.T) {
	entries := []api.DictEntry{
		entryWith("good", "adjective", []string{"good", "better", "best"},
			sseqOf(senseItem("1", "{bc}of a favorable character or tendency"))),
	}

	var b strings.Builder
	PrintDefinitions(&b, "best", entries, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, `Showing results for "good"`) {
		t.Errorf("stem fallback note missing for irregular form:\n%s", out)
	}
	if !strings.Contains(out, "favorable character") {
		t.Errorf("definition missing:\n%s", out)
	}
}

func TestPrintSyn_StemFallback_Happiest(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntryWithStems("happy", "adjective",
			[][]string{{"glad", "joyful"}},
			[][]string{{"sad", "unhappy"}},
			[]string{"happy", "happier", "happiest", "happily"},
		),
	}

	var b strings.Builder
	PrintSynonyms(&b, "happiest", entries, FormatPlain)
	out := b.String()

	if !strings.Contains(out, `Showing results for "happy"`) {
		t.Errorf("stem fallback note missing:\n%s", out)
	}
	if !strings.Contains(out, "glad") {
		t.Errorf("synonyms missing after stem fallback:\n%s", out)
	}
}

func TestPrintAnt_StemFallback_Sadder(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntryWithStems("sad", "adjective",
			nil,
			[][]string{{"happy", "joyful"}},
			[]string{"sad", "sadder", "saddest"},
		),
	}

	var b strings.Builder
	PrintAntonyms(&b, "sadder", entries, FormatPlain)
	out := b.String()

	if !strings.Contains(out, `Showing results for "sad"`) {
		t.Errorf("stem fallback note missing:\n%s", out)
	}
	if !strings.Contains(out, "happy") {
		t.Errorf("antonyms missing after stem fallback:\n%s", out)
	}
}

// ── maxDefs capping ──────────────────────────────────────────────────────────

func TestPrintDef_MaxDefs_WithPseq(t *testing.T) {
	// Ensure maxDefs counts correctly even when senses come from a pseq block.
	entry := entryWith("set", "verb", nil,
		sseqOf(
			pseqItem(
				senseItem("1 a", "{bc}to put in a specified place"),
				senseItem("b", "{bc}to put in a particular condition"),
				senseItem("c", "{bc}to fix or establish"),
			),
			senseItem("2", "{bc}to become solid"),
			senseItem("3", "{bc}to pass below the horizon"),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "set", []api.DictEntry{entry}, 2, FormatPlain)
	out := b.String()

	count := strings.Count(out, "\n  ") // each def line starts with "  N."
	if count != 2 {
		t.Errorf("expected 2 definitions with maxDefs=2, got %d:\n%s", count, out)
	}
}

// ── multiple def blocks (verb dividers) ─────────────────────────────────────

func TestPrintDef_MultipleDefBlocks(t *testing.T) {
	// Some verbs have separate def blocks for transitive and intransitive uses,
	// each with their own sseq. Both blocks' senses should appear.
	entry := entryMultiBlock("run", "verb",
		api.DefBlock{Vd: "intransitive verb", Sseq: sseqOf(
			senseItem("1", "{bc}to go at a pace faster than a walk"),
		)},
		api.DefBlock{Vd: "transitive verb", Sseq: sseqOf(
			senseItem("1", "{bc}to cause to run"),
		)},
	)

	var b strings.Builder
	PrintDefinitions(&b, "run", []api.DictEntry{entry}, 10, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "pace faster than a walk") {
		t.Errorf("intransitive verb sense missing:\n%s", out)
	}
	if !strings.Contains(out, "cause to run") {
		t.Errorf("transitive verb sense missing:\n%s", out)
	}
}

// ── no fl / empty def guards ─────────────────────────────────────────────────

func TestPrintDef_NoFlSkipped(t *testing.T) {
	entries := []api.DictEntry{
		{Hwi: api.Hwi{Hw: "test"}, Fl: ""},                          // no fl → skip
		entryWith("test", "noun", nil,                                 // normal entry
			sseqOf(senseItem("1", "{bc}a proper entry"))),
	}

	var b strings.Builder
	PrintDefinitions(&b, "test", entries, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "proper entry") {
		t.Errorf("valid entry missing:\n%s", out)
	}
}

func TestPrintDef_AllSensesEmpty_ShowsNoDefsFound(t *testing.T) {
	// Entry where every sense resolves to empty text after markup stripping.
	entry := entryWith("glitch", "noun", nil,
		sseqOf(
			senseItem("1", "{it}{/it}"),
			senseItem("2", ""),
		),
	)

	var b strings.Builder
	PrintDefinitions(&b, "glitch", []api.DictEntry{entry}, 5, FormatPlain)
	if !strings.Contains(b.String(), "No definitions found") {
		t.Errorf("expected 'No definitions found' when all senses empty:\n%s", b.String())
	}
}

// ── thesaurus: multiple sense groups ─────────────────────────────────────────

func TestPrintSyn_FastMultipleSenseGroups(t *testing.T) {
	// "fast" has synonyms grouped by different meanings (speed vs. firmness).
	entries := []api.ThesEntry{
		makeThesEntry("fast", "adjective",
			[][]string{
				{"quick", "speedy", "swift"},
				{"fixed", "firm", "secure"},
			},
			nil,
		),
	}

	var b strings.Builder
	PrintSynonyms(&b, "fast", entries, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "Sense 1:") {
		t.Errorf("Sense 1 label missing:\n%s", out)
	}
	if !strings.Contains(out, "Sense 2:") {
		t.Errorf("Sense 2 label missing:\n%s", out)
	}
	if !strings.Contains(out, "quick") {
		t.Errorf("Sense 1 synonyms missing:\n%s", out)
	}
	if !strings.Contains(out, "fixed") {
		t.Errorf("Sense 2 synonyms missing:\n%s", out)
	}
}

func TestPrintAnt_FastMultipleSenseGroups(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("fast", "adjective",
			nil,
			[][]string{
				{"slow", "sluggish"},
				{"loose", "insecure"},
			},
		),
	}

	var b strings.Builder
	PrintAntonyms(&b, "fast", entries, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "Sense 1:") {
		t.Errorf("Sense 1 label missing:\n%s", out)
	}
	if !strings.Contains(out, "slow") || !strings.Contains(out, "loose") {
		t.Errorf("antonyms missing:\n%s", out)
	}
}

// ── multi-word phrase queries ─────────────────────────────────────────────────

func TestPrintDef_MultiWord_SpillTheBeans(t *testing.T) {
	entries := []api.DictEntry{
		entryWith("spill the beans", "verb phrase", []string{"spill the beans"},
			sseqOf(senseItem("1", "{bc}to reveal secret information prematurely"))),
		// Unrelated single-word entry should be filtered.
		entryWith("spill", "verb", nil,
			sseqOf(senseItem("1", "{bc}to cause to fall or flow"))),
	}

	var b strings.Builder
	PrintDefinitions(&b, "spill the beans", entries, 5, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "reveal secret information") {
		t.Errorf("phrase definition missing:\n%s", out)
	}
	if strings.Contains(out, "fall or flow") {
		t.Errorf("unrelated single-word entry should be filtered:\n%s", out)
	}
}

func TestPrintSyn_MultiWord_SpillTheBeans(t *testing.T) {
	entries := []api.ThesEntry{
		makeThesEntry("spill the beans", "verb phrase",
			[][]string{{"let the cat out of the bag", "give away", "blab"}},
			nil,
		),
		makeThesEntry("spill", "verb", [][]string{{"pour", "shed"}}, nil),
	}

	var b strings.Builder
	PrintSynonyms(&b, "spill the beans", entries, FormatPlain)
	out := b.String()

	if !strings.Contains(out, "let the cat out of the bag") {
		t.Errorf("phrase synonym missing:\n%s", out)
	}
	if strings.Contains(out, "pour") {
		t.Errorf("unrelated single-word synonyms should be filtered:\n%s", out)
	}
}

// ── extractSenseItem: unit tests ─────────────────────────────────────────────

func TestExtractSenseItem_Sense(t *testing.T) {
	senses := extractSenseItem(senseItem("1", "a definition"))
	if len(senses) != 1 || senses[0].text != "a definition" {
		t.Errorf("got %+v", senses)
	}
}

func TestExtractSenseItem_Bs(t *testing.T) {
	senses := extractSenseItem(bsItem("1", "a bs sense"))
	if len(senses) != 1 || senses[0].text != "a bs sense" {
		t.Errorf("got %+v", senses)
	}
}

func TestExtractSenseItem_Sen(t *testing.T) {
	// sen has no dt → no senses.
	senses := extractSenseItem(senItem("1"))
	if len(senses) != 0 {
		t.Errorf("sen should yield no senses, got %+v", senses)
	}
}

func TestExtractSenseItem_Pseq(t *testing.T) {
	senses := extractSenseItem(pseqItem(
		senseItem("1 a", "first sub"),
		senseItem("b", "second sub"),
	))
	if len(senses) != 2 {
		t.Errorf("expected 2 senses from pseq, got %d: %+v", len(senses), senses)
	}
}

func TestExtractSenseItem_UnknownTag(t *testing.T) {
	senses := extractSenseItem([]any{"snote", map[string]any{"t": "a supplemental note"}})
	if len(senses) != 0 {
		t.Errorf("unknown tag should yield no senses, got %+v", senses)
	}
}

func TestExtractSenseItem_ShortItem(t *testing.T) {
	senses := extractSenseItem([]any{"sense"}) // only 1 element, no data
	if len(senses) != 0 {
		t.Errorf("short item should yield no senses, got %+v", senses)
	}
}
