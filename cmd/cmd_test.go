package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/mweb/internal/api"
	"github.com/user/mweb/internal/config"
)

// executeCmd runs the root command with the given args. cfgOverride, if non-nil,
// is injected directly (skipping config file loading). Returns captured stdout
// output and any error.
func executeCmd(cfgOverride *config.Config, args []string) (string, error) {
	cfg = cfgOverride
	jsonOutput = false

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), err
}

// setupMockServer starts an httptest server that always returns body, injects the
// newAPIClient factory to point at it, and returns a cleanup func.
func setupMockServer(t *testing.T, body string) (*httptest.Server, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))

	origFactory := newAPIClient
	newAPIClient = func(dictKey, thesKey string) *api.Client {
		c := api.NewClient(dictKey, thesKey)
		c.SetBaseURL(srv.URL)
		return c
	}

	return srv, func() {
		srv.Close()
		newAPIClient = origFactory
	}
}

func testCfg() *config.Config {
	return &config.Config{
		APIKeyDict:      "dict-key",
		APIKeyThesaurus: "thes-key",
		OutputFormat:    "plain",
		MaxDefinitions:  5,
	}
}

// ---- def command ----

func TestDef_MissingAPIKey(t *testing.T) {
	_, err := executeCmd(&config.Config{APIKeyDict: "", MaxDefinitions: 5}, []string{"def", "test"})
	if err == nil {
		t.Fatal("expected error for missing dict API key")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("error should mention API key, got: %v", err)
	}
}

func TestDef_WrongArgCount(t *testing.T) {
	_, err := executeCmd(testCfg(), []string{"def"})
	if err == nil {
		t.Fatal("expected error for missing word argument")
	}
}

func TestDef_TooManyArgs(t *testing.T) {
	_, err := executeCmd(testCfg(), []string{"def", "word1", "word2"})
	if err == nil {
		t.Fatal("expected error for too many arguments")
	}
}

func TestDef_WordNotFound(t *testing.T) {
	_, cleanup := setupMockServer(t, `["running","runner"]`)
	defer cleanup()

	_, err := executeCmd(testCfg(), []string{"def", "runn"})
	if err == nil {
		t.Fatal("expected error for word not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestDef_Success(t *testing.T) {
	body := `[{"meta":{"id":"test:1","stems":["test"],"offensive":false},"hwi":{"hw":"test","prs":[{"mw":"ˈtest"}]},"fl":"noun","def":[{"sseq":[[["sense",{"sn":"1","dt":[["text","a procedure for testing"]]}]]]}],"et":[],"date":"14th century"}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"def", "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "noun") {
		t.Errorf("expected 'noun' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "a procedure for testing") {
		t.Errorf("expected definition text in output, got:\n%s", out)
	}
}

// ---- syn command ----

func TestSyn_MissingAPIKey(t *testing.T) {
	_, err := executeCmd(&config.Config{APIKeyThesaurus: "", MaxDefinitions: 5}, []string{"syn", "happy"})
	if err == nil {
		t.Fatal("expected error for missing thesaurus API key")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("error should mention API key, got: %v", err)
	}
}

func TestSyn_Success(t *testing.T) {
	body := `[{"meta":{"id":"happy","syns":[["glad","joyful"]],"ants":[],"offensive":false},"hwi":{"hw":"happy"},"fl":"adjective","def":[]}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"syn", "happy"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "glad") {
		t.Errorf("expected 'glad' in synonyms output, got:\n%s", out)
	}
}

// ---- ant command ----

func TestAnt_MissingAPIKey(t *testing.T) {
	_, err := executeCmd(&config.Config{APIKeyThesaurus: "", MaxDefinitions: 5}, []string{"ant", "happy"})
	if err == nil {
		t.Fatal("expected error for missing thesaurus API key")
	}
}

func TestAnt_Success(t *testing.T) {
	body := `[{"meta":{"id":"happy","syns":[],"ants":[["sad","unhappy"]],"offensive":false},"hwi":{"hw":"happy"},"fl":"adjective","def":[]}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"ant", "happy"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sad") {
		t.Errorf("expected 'sad' in antonyms output, got:\n%s", out)
	}
}

// ---- multi-word phrase queries ----

func TestDef_MultiWordPhrase(t *testing.T) {
	body := `[{"meta":{"id":"spill the beans","stems":["spill the beans"],"offensive":false},"hwi":{"hw":"spill the beans"},"fl":"verb phrase","def":[{"sseq":[[["sense",{"sn":"1","dt":[["text","to reveal secret information"]]}]]]}],"et":[],"date":""}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"def", "spill", "the", "beans"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "reveal secret information") {
		t.Errorf("expected phrase definition in output, got:\n%s", out)
	}
}

func TestSyn_MultiWordPhrase(t *testing.T) {
	body := `[{"meta":{"id":"spill the beans","syns":[["let the cat out of the bag"]],"ants":[],"offensive":false},"hwi":{"hw":"spill the beans"},"fl":"verb phrase","def":[]}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"syn", "spill", "the", "beans"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "let the cat out of the bag") {
		t.Errorf("expected phrase synonym in output, got:\n%s", out)
	}
}

func TestAnt_MultiWordPhrase(t *testing.T) {
	body := `[{"meta":{"id":"spill the beans","syns":[],"ants":[["keep secret"]],"offensive":false},"hwi":{"hw":"spill the beans"},"fl":"verb phrase","def":[]}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"ant", "spill", "the", "beans"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "keep secret") {
		t.Errorf("expected phrase antonym in output, got:\n%s", out)
	}
}

func TestDef_NoArgsError(t *testing.T) {
	_, err := executeCmd(testCfg(), []string{"def"})
	if err == nil {
		t.Fatal("expected error for missing word argument")
	}
}

// ---- use command ----

func TestUse_MissingAPIKey(t *testing.T) {
	_, err := executeCmd(&config.Config{APIKeyDict: "", MaxDefinitions: 5}, []string{"use", "test"})
	if err == nil {
		t.Fatal("expected error for missing dict API key")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("error should mention API key, got: %v", err)
	}
}

func TestUse_WrongArgCount(t *testing.T) {
	_, err := executeCmd(testCfg(), []string{"use"})
	if err == nil {
		t.Fatal("expected error for missing word argument")
	}
}

func TestUse_WordNotFound(t *testing.T) {
	_, cleanup := setupMockServer(t, `["running","runner"]`)
	defer cleanup()

	_, err := executeCmd(testCfg(), []string{"use", "runn"})
	if err == nil {
		t.Fatal("expected error for word not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestUse_Success(t *testing.T) {
	body := `[{"meta":{"id":"test:1","stems":["test"],"offensive":false},"hwi":{"hw":"test"},"fl":"noun","def":[{"sseq":[[["sense",{"sn":"1","dt":[["text","{bc}a means of testing"],["vis",[{"t":"the {it}test{/it} was easy"}]]]}]]]}]}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"use", "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "test (noun)") {
		t.Errorf("expected header 'test (noun)', got:\n%s", out)
	}
	if !strings.Contains(out, "the test was easy") {
		t.Errorf("expected example with markup stripped, got:\n%s", out)
	}
}

func TestUse_NoExamples(t *testing.T) {
	// Entry exists but has no vis items in dt
	body := `[{"meta":{"id":"test:1","stems":["test"],"offensive":false},"hwi":{"hw":"test"},"fl":"noun","def":[{"sseq":[[["sense",{"sn":"1","dt":[["text","{bc}a means of testing"]]}]]]}]}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"use", "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No example usage found") {
		t.Errorf("expected 'No example usage found', got:\n%s", out)
	}
}

func TestUse_MultiWordPhrase(t *testing.T) {
	body := `[{"meta":{"id":"spill the beans","stems":["spill the beans"],"offensive":false},"hwi":{"hw":"spill the beans"},"fl":"verb phrase","def":[{"sseq":[[["sense",{"sn":"1","dt":[["text","{bc}to reveal secret information"],["vis",[{"t":"he spilled the beans about the party"}]]]}]]]}]}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"use", "spill", "the", "beans"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "spill the beans") {
		t.Errorf("expected phrase in output, got:\n%s", out)
	}
	if !strings.Contains(out, "he spilled the beans about the party") {
		t.Errorf("expected example sentence, got:\n%s", out)
	}
}

func TestUse_JSONFlag(t *testing.T) {
	body := `[{"meta":{"id":"test:1","stems":["test"],"offensive":false},"hwi":{"hw":"test"},"fl":"noun","def":[{"sseq":[[["sense",{"sn":"1","dt":[["text","{bc}a means of testing"],["vis",[{"t":"a simple test"}]]]}]]]}]}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"--json", "use", "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected JSON array output, got:\n%s", out)
	}
}

// ---- --json flag ----

func TestDef_JSONFlag(t *testing.T) {
	body := `[{"meta":{"id":"run:1","stems":["run"],"offensive":false},"hwi":{"hw":"run"},"fl":"verb","def":[],"et":[],"date":""}]`
	_, cleanup := setupMockServer(t, body)
	defer cleanup()

	out, err := executeCmd(testCfg(), []string{"--json", "def", "run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected JSON array output, got:\n%s", out)
	}
}
