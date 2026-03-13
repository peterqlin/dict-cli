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
